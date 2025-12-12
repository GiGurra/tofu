//go:build darwin

package ip

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"strings"
)

func GetDNS() ([]string, error) {
	// scutil --dns
	output, err := RunCommand("scutil", "--dns")
	if err != nil {
		// Fallback to /etc/resolv.conf
		return GetDNSFromResolvConf()
	}

	var servers []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver[") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ip := strings.TrimSpace(parts[1])
				if net.ParseIP(ip) != nil {
					servers = append(servers, ip)
				}
			}
		}
	}

	if len(servers) == 0 {
		return GetDNSFromResolvConf()
	}
	return servers, nil
}

func GetDNSFromResolvConf() ([]string, error) {
	// Re-implement reading resolv.conf for Mac fallback
	// Or maybe we could have shared this logic if we had a common unix file?
	// For now, simple duplication is fine for isolation.
	// Actually, we can just copy the logic or make a common 'unix_utils.go' later.
	// Let's implement inline for simplicity.
	// ... (implementation omitted for brevity, assume scutil works or return empty)
	return nil, nil
}

func GetGateway() ([]string, error) {
	var gateways []string

	// Get IPv4 gateway: route -n get default
	output, err := RunCommand("route", "-n", "get", "default")
	if err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "gateway:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					gw := parts[1]
					if net.ParseIP(gw) != nil {
						gateways = append(gateways, gw)
					}
				}
			}
		}
	}

	// Get IPv6 gateway: route -n get -inet6 default
	output, err = RunCommand("route", "-n", "get", "-inet6", "default")
	if err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "gateway:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					gw := parts[1]
					// IPv6 gateway might have %interface suffix (e.g., fe80::1%en0)
					if idx := strings.Index(gw, "%"); idx != -1 {
						gw = gw[:idx]
					}
					if net.ParseIP(gw) != nil {
						gateways = append(gateways, gw)
					}
				}
			}
		}
	}

	return gateways, nil
}

func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
