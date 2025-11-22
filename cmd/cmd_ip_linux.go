//go:build linux

package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func getDNS() ([]string, error) {
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	var servers []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				servers = append(servers, parts[1])
			}
		}
	}
	return servers, nil
}

func getGateway() (string, error) {
	content, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 8 {
			// Destination is 2nd field, Mask is 8th field
			// Default gateway has Destination 00000000 and Mask 00000000
			if fields[1] == "00000000" && fields[7] == "00000000" {
				// Gateway IP is 3rd field (hex)
				gwHex := fields[2]
				return parseHexIP(gwHex)
			}
		}
	}
	return "", nil
}

func parseHexIP(hexStr string) (string, error) {
	if len(hexStr) != 8 {
		return "", fmt.Errorf("invalid hex IP length")
	}
	var bytes [4]byte
	_, err := fmt.Sscanf(hexStr, "%02x%02x%02x%02x", &bytes[3], &bytes[2], &bytes[1], &bytes[0])
	if err != nil {
		return "", err
	}
	return net.IPv4(bytes[0], bytes[1], bytes[2], bytes[3]).String(), nil
}
