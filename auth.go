package main

import (
	"fmt"
	"net"
)

func selectMethod(buffer []byte, auth bool) byte {
	methodLen := int(buffer[0])
	if methodLen > len(buffer)-1 {
		methodLen = len(buffer) - 1
	}

	var method byte = 0xff
	for i := 1; i <= methodLen; i++ {
		if (auth && buffer[i] == 2) || (!auth && buffer[i] == 0) {
			method = buffer[i]
			break
		}
	}

	return method
}

func authUser(conn *net.TCPConn, username string, password string) bool {
	buffer := make([]byte, 513)
	n, err := conn.Read(buffer)

	if err != nil || n < 5 {
		return false
	}

	buffer = buffer[:n]

	if buffer[0] != 1 || n < int(buffer[1]+4) || n < int(buffer[1]+buffer[int(2+buffer[1])])+3 {
		fmt.Println("[WARNING] Malformed authentication message:", buffer)
		return false
	}

	if string(buffer[2:2+buffer[1]]) == username && string(buffer[3+buffer[1]:3+buffer[1]+buffer[2+buffer[1]]]) == password {
		_, err = conn.Write([]byte{0x1, 0x0})
		if err != nil {
			conn.Close()
			return false
		}
	} else {
		conn.Write([]byte{0x1, 0xff})
		conn.Close()
		return false
	}

	return true
}
