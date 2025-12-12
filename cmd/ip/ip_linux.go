//go:build linux

package ip

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetDNS() ([]string, error) {
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

func GetGateway() ([]string, error) {
	var gateways []string

	// Get IPv4 gateway from /proc/net/route
	content, err := os.ReadFile("/proc/net/route")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 8 {
				// Destination is 2nd field, Mask is 8th field
				// Default gateway has Destination 00000000 and Mask 00000000
				if fields[1] == "00000000" && fields[7] == "00000000" {
					// Gateway IP is 3rd field (hex)
					gwHex := fields[2]
					gw, err := ParseHexIP(gwHex)
					if err == nil && gw != "" {
						gateways = append(gateways, gw)
					}
				}
			}
		}
	}

	// Get IPv6 gateway from /proc/net/ipv6_route
	content, err = os.ReadFile("/proc/net/ipv6_route")
	if err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			// Format: dest dest_prefix source src_prefix nexthop metric refcnt use flags iface
			// Default route has dest 00000000000000000000000000000000 and prefix 00
			if len(fields) >= 10 {
				dest := fields[0]
				prefix := fields[1]
				nexthop := fields[4]
				// Default route: ::/0 (all zeros dest with 00 prefix length)
				if dest == "00000000000000000000000000000000" && prefix == "00" {
					// Skip if nexthop is all zeros (no gateway)
					if nexthop != "00000000000000000000000000000000" {
						gw := ParseHexIPv6(nexthop)
						if gw != "" {
							gateways = append(gateways, gw)
						}
					}
				}
			}
		}
	}

	return gateways, nil
}

func ParseHexIP(hexStr string) (string, error) {
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

func ParseHexIPv6(hexStr string) string {
	if len(hexStr) != 32 {
		return ""
	}
	var ip net.IP = make([]byte, 16)
	for i := 0; i < 16; i++ {
		var b byte
		_, err := fmt.Sscanf(hexStr[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return ""
		}
		ip[i] = b
	}
	return ip.String()
}
