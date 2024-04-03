package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	ATYP_IPV4   = 1
	ATYP_DOMAIN = 3
	ATYP_IPV6   = 4
)

const (
	CMD_CONNECT       = 1
	CMD_BIND          = 2
	CMD_UDP_ASSOCIATE = 3
)

const (
	REP_SUCCEEDED                  = 0
	REP_GENERAL_FAILURE            = 1
	REP_CONNECTION_REFUSED         = 5
	REP_COMMAND_NOT_SUPPORTED      = 7
	REP_ADDRESS_TYPE_NOT_SUPPORTED = 8
)

var tcpAddress = "0.0.0.0:1080"
var username = ""
var password = ""
var udpRelayAddress = ""
var udpRelayAddressIpv4 net.IP = nil
var udpRelayAddressType byte = 0

func isValidAddress() bool {
	/* Valid IPv4 */
	udpRelayAddressIpv4 = net.ParseIP(udpRelayAddress)
	if udpRelayAddressIpv4 != nil && udpRelayAddressIpv4.To4() != nil {
		udpRelayAddressType = ATYP_IPV4
		return true
	}

	/* Valid Domain */
	if _, err := net.LookupHost(tcpAddress); err == nil {
		udpRelayAddressType = ATYP_DOMAIN
		return true
	}

	return false
}

func parseCmd() {
	/* UDP relay address is required */
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [UDP Relay Address] <username> <password> <TCP Listening Address>\n", os.Args[0])
		os.Exit(-1)
	}

	/* Check if the address is a valid IPv4 address or domain */
	udpRelayAddress = os.Args[1]
	if !isValidAddress() {
		fmt.Println("[FATAL] Invalid IP", os.Args[1])
		os.Exit(-1)
	}

	if len(os.Args) > 3 {
		username = os.Args[2]
		password = os.Args[3]
	}

	if len(os.Args) > 4 {
		tcpAddress = os.Args[4]
	}
}

func parseAddress(buffer []byte) (string, int) {
	address := ""
	bytesParsed := 0

	/* Construct address string */
	if buffer[0] == ATYP_IPV4 {
		address += net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]).String()
		bytesParsed = 7
	} else if buffer[0] == ATYP_DOMAIN {
		if len(buffer) < int(buffer[1])+4 {
			return address, 0
		}
		address += string(buffer[2 : 2+buffer[1]])
		bytesParsed = int(buffer[1]) + 4
	} else {
		return address, 0
	}

	/* Construct port string */
	return address + ":" + strconv.FormatUint(uint64(binary.BigEndian.Uint16(buffer[bytesParsed-2:])), 10), bytesParsed
}

func ip2Int(ip net.IP) uint32 {
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
