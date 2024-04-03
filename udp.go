package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func handlePackets(tcp *net.TCPConn, udp *net.UDPConn, clientIp net.IP) {
	var clientAddress *net.UDPAddr = nil
	udpIpv4DomainMap := make(map[uint32]string)

	for {
		buffer := make([]byte, 4096)
		n, peer, err := udp.ReadFromUDP(buffer)
		if err != nil || n < 11 {
			if err == nil {
				fmt.Println("[ERROR] UDP request is too short!")
			} else if !errors.Is(err, net.ErrClosed) {
				fmt.Println("[ERROR]", err)
			}
			tcp.Close()
			return
		}

		buffer = buffer[:n]

		if peer.IP.Equal(clientIp) {
			/* We do not support fragmentation, drop the packet */
			if buffer[2] != 0 {
				fmt.Println("[ERROR] Fragment field set", buffer[2])
				tcp.Close()
				return
			}

			/* Initialize client address upon first receive */
			if clientAddress == nil {
				clientAddress = &net.UDPAddr{
					IP:   clientIp,
					Port: peer.Port,
				}
			}

			/* Packet from client, need to decode and then forward */
			domainAddress := buffer[3] == ATYP_DOMAIN
			address, bytesParsed := parseAddress(buffer[3:])
			buffer = buffer[bytesParsed+3:]

			var udpAddress *net.UDPAddr
			udpAddress, err = net.ResolveUDPAddr("udp4", address)
			if err != nil || bytesParsed == 0 {
				fmt.Println("[ERROR] Cannot parse UDP address", err)
				tcp.Close()
				return
			}

			_, err = udp.WriteToUDP(buffer, udpAddress)
			if err != nil {
				fmt.Println("[ERROR] Cannot forward UDP data to target host", err)
				tcp.Close()
				return
			}

			/* Add Domain IPv4 mapping to the concurrent map */
			if domainAddress {
				udpIpv4DomainMap[ip2Int(udpAddress.IP)] = strings.Split(address, ":")[0]
			}
		} else {
			/* Packet from server, need to encode and forward */
			replySize := n + 6
			domain, domainAddress := udpIpv4DomainMap[ip2Int(peer.IP)]

			if domainAddress {
				replySize += len(domain) + 1
			} else {
				replySize += 4
			}

			reply := make([]byte, replySize)

			if domainAddress {
				reply[3] = ATYP_DOMAIN
				reply[4] = byte(len(domain))
				copy(reply[5:], domain)
			} else {
				reply[3] = ATYP_IPV4
				reply[4] = peer.IP.To4()[0]
				reply[5] = peer.IP.To4()[1]
				reply[6] = peer.IP.To4()[2]
				reply[7] = peer.IP.To4()[3]
			}

			portIndex := replySize - n - 2

			/* Copy port value */
			reply[portIndex] = byte(peer.Port >> 8)
			reply[portIndex+1] = byte(peer.Port & 0xff)

			/* Copy payload */
			copy(reply[portIndex+2:], buffer)

			_, err = udp.WriteTo(reply, clientAddress)
			if err != nil {
				fmt.Println("[ERROR] Cannot forward UDP data to client", err)
				tcp.Close()
				return
			}
		}
	}
}

func handleUdp(tcp *net.TCPConn, buffer []byte) {
	/* Get client address */
	clientAddr := tcp.RemoteAddr().(*net.TCPAddr)

	/* Bind to a UDP port */
	udpAddr, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:0")
	udp, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		fmt.Println("[ERROR] Cannot create UDP connection", err)
		replyWithCode(tcp, buffer, REP_GENERAL_FAILURE)
		return
	}

	/* Reply with server address and port */
	buffer[1] = 0
	buffer[3] = udpRelayAddressType

	/* Copy address to the reply */
	if udpRelayAddressType == ATYP_IPV4 {
		/* If address type is IPv4, just copy to the buffer and resize slice */
		ip := udpRelayAddressIpv4.To4()
		buffer[4] = ip[0]
		buffer[5] = ip[1]
		buffer[6] = ip[2]
		buffer[7] = ip[3]

		/* Resize the buffer slice */
		buffer = buffer[:10]
	} else {
		/* If address type is domain, resize the slice and then copy */
		tmp := make([]byte, 7+len(udpRelayAddress))
		copy(tmp, buffer)
		buffer = tmp
		buffer[4] = byte(len(udpRelayAddress))
		copy(buffer[5:], udpRelayAddress)
	}

	/* Copy port to the reply */
	buffer[len(buffer)-2] = byte(udp.LocalAddr().(*net.UDPAddr).Port >> 8)
	buffer[len(buffer)-1] = byte(udp.LocalAddr().(*net.UDPAddr).Port & 0xff)

	/* Send the reply */
	_, err = tcp.Write(buffer)
	if err != nil {
		fmt.Println("[ERROR] Cannot send UDP relay reply")
		udp.Close()
		tcp.Close()
		return
	}

	/* Handle packets in a separate goroutine as we need to wait TCP connection */
	go handlePackets(tcp, udp, clientAddr.IP.To4())

	/* Wait utils TCP connection is terminated */
	for err == nil {
		buffer = make([]byte, 8)
		_, err = tcp.Read(buffer)
	}

	/* When TCP connection is closed, the UDP ASSOCIATE ends */
	udp.Close()
}
