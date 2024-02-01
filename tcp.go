package main

import (
	"fmt"
	"net"
)

func replyWithCode(conn *net.TCPConn, buffer []byte, code byte) {
	buffer[1] = code
	conn.Write(buffer)
	conn.Close()
}

func proxy(src *net.TCPConn, dst *net.TCPConn) {
	defer dst.Close()
	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}

		n, err = dst.Write(buffer[:n])
		if err != nil {
			return
		}
	}

	/* For some reason, the io.Copy() that uses splice() on linux has worse performance */
	//var n int64 = 1
	//var err error = nil
	//for n != 0 && err == nil {
	//	n, err = io.Copy(dst, src)
	//}
	//dst.Close()
}

func handleTcp(conn *net.TCPConn) {
	/* Maximum size of method selection message is 1 + 1 + 255 = 257 */
	buffer := make([]byte, 257)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("[ERROR] Cannot read method selection message")
		conn.Close()
		return
	}
	buffer = buffer[:n]

	/* We only supports SOCKS5 */
	if n == 0 {
		if buffer[0] != 5 {
			fmt.Println("[WARNING] Unsupported SOCKS version", buffer[0])
		}
		conn.Close()
		return
	}

	_, err = conn.Write([]byte{0x5, 0x0})
	if err != nil {
		fmt.Println("[ERROR] Cannot send the method selection reply", err.Error())
		conn.Close()
		return
	}

	/* Process SOCKS5 request */
	buffer = make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil || n < 10 {
		fmt.Println("[ERROR] Cannot read request with length", n)
		conn.Close()
		return
	}
	buffer = buffer[:n]

	/* Validate CMD */
	if buffer[1] != CMD_CONNECT && buffer[1] != CMD_UDP_ASSOCIATE {
		fmt.Println("[ERROR] Invalid CMD", buffer[1], "from", conn.RemoteAddr().String())
		replyWithCode(conn, buffer, REP_COMMAND_NOT_SUPPORTED)
		return
	}

	/* Construct address string */
	var address string
	if address, _ = parseAddress(buffer[3:]); address == "" {
		fmt.Println("[ERROR] Invalid address", buffer, "from", conn.RemoteAddr().String())
		replyWithCode(conn, buffer, REP_ADDRESS_TYPE_NOT_SUPPORTED)
		return
	}

	/* UDP associate */
	if buffer[1] == CMD_UDP_ASSOCIATE {
		go handleUdp(conn, buffer)
		return
	}

	/* Connect the remote host */
	dstConn, err := net.Dial("tcp4", address)
	if err != nil {
		fmt.Println("[ERROR]", err.Error())
		replyWithCode(conn, buffer, REP_CONNECTION_REFUSED)
		return
	}

	/* Connection established */
	buffer[1] = REP_SUCCEEDED
	_, err = conn.Write(buffer)
	if err != nil {
		fmt.Println("[ERROR] Cannot send reply message", err.Error())
		conn.Close()
		dstConn.Close()
		return
	}

	/* Splice data between two conn */
	go proxy(conn, dstConn.(*net.TCPConn))
	go proxy(dstConn.(*net.TCPConn), conn)
}
