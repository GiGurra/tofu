package ip

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	LocalOnly bool `short:"l" help:"Only show local interfaces, do not attempt to discover public IP."`
	Json      bool `short:"j" help:"Output in JSON format."`
}

type IPOutput struct {
	Interfaces     map[string][]string `json:"interfaces"`
	PublicIP       string              `json:"public_ip,omitempty"`
	PublicIPError  string              `json:"public_ip_error,omitempty"`
	DNSServers     []string            `json:"dns_servers,omitempty"`
	DNSError       string              `json:"dns_error,omitempty"`
	Gateways       []string            `json:"gateways,omitempty"`
	GatewaysError  string              `json:"gateways_error,omitempty"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "ip",
		Short:       "Show local and public IP addresses",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			runIp(params, os.Stdout)
		},
	}.ToCobra()
}

func runIp(params *Params, stdout io.Writer) {
	output := IPOutput{
		Interfaces: make(map[string][]string),
	}

	// Local Interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		if params.Json {
			output.Interfaces = nil
			outputJSON(stdout, output)
		} else {
			fmt.Fprintf(stdout, "Error getting interfaces: %v\n", err)
		}
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		if len(addrs) > 0 {
			var addrStrings []string
			for _, addr := range addrs {
				addrStrings = append(addrStrings, addr.String())
			}
			output.Interfaces[i.Name] = addrStrings
		}
	}

	if !params.LocalOnly {
		ip, err := GetPublicIP()
		if err != nil {
			output.PublicIPError = err.Error()
		} else {
			output.PublicIP = strings.TrimSpace(ip)
		}
	}

	// DNS Servers
	dns, err := GetDNS()
	if err != nil {
		output.DNSError = err.Error()
	} else {
		output.DNSServers = dns
	}

	// Default Gateway
	gateways, err := GetGateway()
	if err != nil {
		output.GatewaysError = err.Error()
	} else {
		output.Gateways = gateways
	}

	if params.Json {
		outputJSON(stdout, output)
	} else {
		outputPlain(stdout, params, output)
	}
}

func outputJSON(stdout io.Writer, output IPOutput) {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(output)
}

func outputPlain(stdout io.Writer, params *Params, output IPOutput) {
	fmt.Fprintln(stdout, "Local Interfaces:")
	for name, addrs := range output.Interfaces {
		fmt.Fprintf(stdout, "  %s:\n", name)
		for _, addr := range addrs {
			fmt.Fprintf(stdout, "    %s\n", addr)
		}
	}

	if !params.LocalOnly {
		fmt.Fprintln(stdout, "\nPublic IP:")
		if output.PublicIPError != "" {
			fmt.Fprintf(stdout, "  Error discovering public IP: %s\n", output.PublicIPError)
		} else {
			fmt.Fprintf(stdout, "  %s\n", output.PublicIP)
		}
	}

	fmt.Fprintln(stdout, "\nDNS Servers:")
	if output.DNSError != "" {
		fmt.Fprintf(stdout, "  Error getting DNS: %s\n", output.DNSError)
	} else if len(output.DNSServers) > 0 {
		for _, d := range output.DNSServers {
			fmt.Fprintf(stdout, "  %s\n", d)
		}
	} else {
		fmt.Fprintln(stdout, "  (none found)")
	}

	fmt.Fprintln(stdout, "\nDefault Gateway:")
	if output.GatewaysError != "" {
		fmt.Fprintf(stdout, "  Error getting Gateway: %s\n", output.GatewaysError)
	} else if len(output.Gateways) > 0 {
		for _, gw := range output.Gateways {
			fmt.Fprintf(stdout, "  %s\n", gw)
		}
	} else {
		fmt.Fprintln(stdout, "  (none found)")
	}
}

func GetPublicIP() (string, error) {
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
