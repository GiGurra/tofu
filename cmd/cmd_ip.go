package cmd

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type IpParams struct {
	LocalOnly bool `short:"l" help:"Only show local interfaces, do not attempt to discover public IP."`
}

func IpCmd() *cobra.Command {
	return boa.CmdT[IpParams]{
		Use:         "ip",
		Short:       "Show local and public IP addresses",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *IpParams, cmd *cobra.Command, args []string) {
			runIp(params, os.Stdout)
		},
	}.ToCobra()
}

func runIp(params *IpParams, stdout io.Writer) {
	// Local Interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(stdout, "Error getting interfaces: %v\n", err)
		return
	}

	fmt.Fprintln(stdout, "Local Interfaces:")
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Fprintf(stdout, "  %s: Error getting addresses: %v\n", i.Name, err)
			continue
		}

		if len(addrs) > 0 {
			fmt.Fprintf(stdout, "  %s:\n", i.Name)
			for _, addr := range addrs {
				// Clean up the address string if needed, but addr.String() is usually fine (e.g. 127.0.0.1/8)
				fmt.Fprintf(stdout, "    %s\n", addr.String())
			}
		}
	}

	if !params.LocalOnly {
		fmt.Fprintln(stdout, "\nPublic IP:")
		ip, err := getPublicIP()
		if err != nil {
			fmt.Fprintf(stdout, "  Error discovering public IP: %v\n", err)
		} else {
			fmt.Fprintf(stdout, "  %s\n", strings.TrimSpace(ip))
		}
	}
}

func getPublicIP() (string, error) {
	// Try a few services in order, return first success
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					return string(body), nil
				}
			}
		}
	}
	return "", fmt.Errorf("failed to reach any public IP discovery service")
}
