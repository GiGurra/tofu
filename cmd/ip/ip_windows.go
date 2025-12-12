//go:build windows

package ip

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"strings"
	"syscall"
)

func GetDNS() ([]string, error) {
	// On Windows, ipconfig /all is a reliable way to get this info without complex syscalls
	output, err := RunCommand("ipconfig", "/all")
	if err != nil {
		return nil, err
	}

	var servers []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for "DNS Servers" or localized variants if possible,
		// but standard "ipconfig" usually outputs "DNS Servers" even in many localized versions,
		// or at least we can search for the pattern.
		// However, accurate parsing of ipconfig is tricky across languages.
		// "DNS Servers . . . . . . . . . . . : 1.1.1.1"
		if strings.Contains(line, "DNS Server") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				ip := strings.TrimSpace(parts[1])
				if net.ParseIP(ip) != nil {
					servers = append(servers, ip)
				}
			}
			// Sometimes there are more on subsequent lines
			for scanner.Scan() {
				subLine := strings.TrimSpace(scanner.Text())
				if subLine == "" {
					// End of section
					break
				}
				// Check if it's a valid IP first (handles IPv6 which contains colons)
				if net.ParseIP(subLine) != nil {
					servers = append(servers, subLine)
				} else if strings.Contains(subLine, ":") {
					// New property line (not an IP), end of DNS section
					break
				}
			}
		}
	}
	return servers, nil
}

func GetGateway() (string, error) {
	// route print 0.0.0.0
	output, err := RunCommand("route", "print", "0.0.0.0")
	if err != nil {
		return "", err
	}

	// Output format:
	// Network Destination        Netmask          Gateway       Interface  Metric
	//           0.0.0.0          0.0.0.0      192.168.1.1    192.168.1.100     25
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
			return fields[2], nil
		}
	}
	return "", nil
}

func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
