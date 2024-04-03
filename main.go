package main

import (
	"fmt"
	"net"
)

func main() {
	/* Get command line arguments */
	parseCmd()

	addr, err := net.ResolveTCPAddr("tcp4", tcpAddress)
	if err != nil {
		fmt.Println("[ERROR] Invalid TCP address", tcpAddress, ":", err)
		return
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Println("[ERROR] Cannot listen on", tcpAddress, ":", err)
		return
	}
	defer listener.Close()

	fmt.Printf("[INFO] Listening on %s...\n", tcpAddress)

	for {
		var conn *net.TCPConn
		conn, err = listener.AcceptTCP()
		if err != nil {
			fmt.Println("[ERROR] Error accepting connection:", err)
			continue
		}
		go handleTcp(conn)
	}
}
