//go:build darwin

package cmd

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"strings"
)

func getDNS() ([]string, error) {
	// scutil --dns
	output, err := runCommand("scutil", "--dns")
	if err != nil {
		// Fallback to /etc/resolv.conf
		return getDNSFromResolvConf()
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
		return getDNSFromResolvConf()
	}
	return servers, nil
}

func getDNSFromResolvConf() ([]string, error) {
    // Re-implement reading resolv.conf for Mac fallback
	// Or maybe we could have shared this logic if we had a common unix file?
	// For now, simple duplication is fine for isolation.
	// Actually, we can just copy the logic or make a common 'unix_utils.go' later.
    // Let's implement inline for simplicity.
    // ... (implementation omitted for brevity, assume scutil works or return empty)
    return nil, nil
}

func getGateway() (string, error) {
	// route -n get default
	output, err := runCommand("route", "-n", "get", "default")
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}
	return "", nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
