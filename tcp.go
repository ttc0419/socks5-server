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

	/* Method selection */
	method := selectMethod(buffer[1:n], username != "" && password != "")
	_, err = conn.Write([]byte{0x5, method})
	if err != nil || method == 0xff {
		if err != nil {
			fmt.Println("[ERROR] Cannot send the method selection reply", err)
		} else {
			fmt.Println("[ERROR] Cannot select method for request", buffer, "from", conn.RemoteAddr().String())
		}
		conn.Close()
		return
	}

	/* Authenticate user if method is username/password */
	if method == 2 && !authUser(conn, username, password) {
		fmt.Println("[WARNING] Invalid username/password attempt from", conn.RemoteAddr().String())
		conn.Close()
		return
	}

	/* Process SOCKS5 request */
	buffer = make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil || n < 10 {
		fmt.Println("[ERROR] Cannot read request with length", n, err)
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
	var dstAddr *net.TCPAddr
	dstAddr, err = net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		fmt.Println("[ERROR] Invalid TCP address", address, err)
		replyWithCode(conn, buffer, REP_ADDRESS_TYPE_NOT_SUPPORTED)
		return
	}

	var dstConn *net.TCPConn
	dstConn, err = net.DialTCP("tcp4", nil, dstAddr)
	if err != nil {
		fmt.Println("[ERROR]", err)
		replyWithCode(conn, buffer, REP_CONNECTION_REFUSED)
		return
	}

	/* Connection established */
	buffer[1] = REP_SUCCEEDED
	_, err = conn.Write(buffer)
	if err != nil {
		fmt.Println("[ERROR] Cannot send reply message", err)
		conn.Close()
		dstConn.Close()
		return
	}

	/* Splice data between two conn */
	go proxy(conn, dstConn)
	go proxy(dstConn, conn)
}
