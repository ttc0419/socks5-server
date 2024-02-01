package main

import (
	"fmt"
	"net"
)

func main() {
	/* Get command line arguments */
	ParseCmd()

	listener, err := net.Listen("tcp4", tcpAddress)
	if err != nil {
		fmt.Println("[ERROR] Cannot listen on", tcpAddress, ":", err)
		return
	}
	defer listener.Close()

	fmt.Printf("[INFO] Listening on %s...\n", tcpAddress)

	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if err != nil {
			fmt.Println("[ERROR] Error accepting connection:", err)
			conn.Close()
			continue
		}
		go handleTcp(conn.(*net.TCPConn))
	}
}
