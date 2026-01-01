package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Hostname string `pos:"true" help:"Hostname to lookup"`
	Server   string `short:"s" help:"DNS server to use (e.g. 8.8.8.8). Defaults to 8.8.8.8:53" default:"8.8.8.8:53"`
	UseOS    bool   `short:"o" long:"use-os" help:"Use OS resolver instead of direct query (ignores --server)"`
	Types    string `short:"t" help:"Comma-separated list of record types to query (A,AAAA,CNAME,MX,TXT,NS,PTR). Default: all common types" default:""`
	Json     bool   `short:"j" help:"Output in JSON format."`
}

type MXRecord struct {
	Pref uint16 `json:"preference"`
	Host string `json:"host"`
}

type DNSOutput struct {
	Server   string     `json:"server"`
	Hostname string     `json:"hostname"`
	A        []string   `json:"a,omitempty"`
	AAAA     []string   `json:"aaaa,omitempty"`
	CNAME    string     `json:"cname,omitempty"`
	MX       []MXRecord `json:"mx,omitempty"`
	TXT      []string   `json:"txt,omitempty"`
	NS       []string   `json:"ns,omitempty"`
	PTR      []string   `json:"ptr,omitempty"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "dns",
		Short:       "Lookup DNS records",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if params.Hostname == "" {
				_ = cmd.Help()
				return
			}
			runDns(params, os.Stdout)
		},
	}.ToCobra()
}

func runDns(params *Params, stdout io.Writer) {
	var resolver *net.Resolver
	var serverName string

	if params.UseOS {
		resolver = net.DefaultResolver
		serverName = "OS Default"
	} else {
		// Ensure port is present
		server := params.Server
		if !strings.Contains(server, ":") {
			server = server + ":53"
		}
		serverName = server

		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				return d.DialContext(ctx, "udp", server)
			},
		}
	}

	output := DNSOutput{
		Server:   serverName,
		Hostname: params.Hostname,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	typesToQuery := parseTypes(params.Types)

	for _, recordType := range typesToQuery {
		switch recordType {
		case "A":
			ips, err := resolver.LookupIPAddr(ctx, params.Hostname)
			if err == nil {
				for _, ip := range ips {
					if ip.IP.To4() != nil {
						output.A = append(output.A, ip.String())
					}
				}
			} else if !params.Json {
				printSection(stdout, "A Records", err, func() {})
			}

		case "AAAA":
			ips, err := resolver.LookupIPAddr(ctx, params.Hostname)
			if err == nil {
				for _, ip := range ips {
					if ip.IP.To4() == nil {
						output.AAAA = append(output.AAAA, ip.String())
					}
				}
			} else if !params.Json {
				printSection(stdout, "AAAA Records", err, func() {})
			}

		case "CNAME":
			cname, err := resolver.LookupCNAME(ctx, params.Hostname)
			if err == nil && cname != "" && cname != params.Hostname && cname != params.Hostname+"." {
				output.CNAME = cname
			} else if err != nil && !params.Json {
				printSection(stdout, "CNAME", err, func() {})
			}

		case "MX":
			mxs, err := resolver.LookupMX(ctx, params.Hostname)
			if err == nil {
				sort.Slice(mxs, func(i, j int) bool { return mxs[i].Pref < mxs[j].Pref })
				for _, mx := range mxs {
					output.MX = append(output.MX, MXRecord{Pref: mx.Pref, Host: mx.Host})
				}
			} else if !params.Json {
				printSection(stdout, "MX Records", err, func() {})
			}

		case "TXT":
			txts, err := resolver.LookupTXT(ctx, params.Hostname)
			if err == nil {
				output.TXT = txts
			} else if !params.Json {
				printSection(stdout, "TXT Records", err, func() {})
			}

		case "NS":
			nss, err := resolver.LookupNS(ctx, params.Hostname)
			if err == nil {
				for _, ns := range nss {
					output.NS = append(output.NS, ns.Host)
				}
			} else if !params.Json {
				printSection(stdout, "NS Records", err, func() {})
			}

		case "PTR":
			names, err := resolver.LookupAddr(ctx, params.Hostname)
			if err == nil {
				output.PTR = names
			} else if !params.Json {
				printSection(stdout, "PTR Records", err, func() {})
			}
		}
	}

	if params.Json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(output)
	} else {
		outputDnsPlain(stdout, params, output)
	}
}

func outputDnsPlain(stdout io.Writer, params *Params, output DNSOutput) {
	fmt.Fprintf(stdout, "Server:  %s\n", output.Server)
	fmt.Fprintf(stdout, "Address: %s\n\n", output.Hostname)

	typesToQuery := parseTypes(params.Types)

	for _, recordType := range typesToQuery {
		switch recordType {
		case "A":
			if len(output.A) > 0 {
				fmt.Fprintln(stdout, "A Records:")
				for _, ip := range output.A {
					fmt.Fprintf(stdout, "  %s\n", ip)
				}
				fmt.Fprintln(stdout)
			}
		case "AAAA":
			if len(output.AAAA) > 0 {
				fmt.Fprintln(stdout, "AAAA Records:")
				for _, ip := range output.AAAA {
					fmt.Fprintf(stdout, "  %s\n", ip)
				}
				fmt.Fprintln(stdout)
			}
		case "CNAME":
			if output.CNAME != "" {
				fmt.Fprintln(stdout, "CNAME:")
				fmt.Fprintf(stdout, "  %s\n", output.CNAME)
				fmt.Fprintln(stdout)
			}
		case "MX":
			if len(output.MX) > 0 {
				fmt.Fprintln(stdout, "MX Records:")
				for _, mx := range output.MX {
					fmt.Fprintf(stdout, "  %d %s\n", mx.Pref, mx.Host)
				}
				fmt.Fprintln(stdout)
			}
		case "TXT":
			if len(output.TXT) > 0 {
				fmt.Fprintln(stdout, "TXT Records:")
				for _, txt := range output.TXT {
					fmt.Fprintf(stdout, "  %s\n", txt)
				}
				fmt.Fprintln(stdout)
			}
		case "NS":
			if len(output.NS) > 0 {
				fmt.Fprintln(stdout, "NS Records:")
				for _, ns := range output.NS {
					fmt.Fprintf(stdout, "  %s\n", ns)
				}
				fmt.Fprintln(stdout)
			}
		case "PTR":
			if len(output.PTR) > 0 {
				fmt.Fprintln(stdout, "PTR Records:")
				for _, name := range output.PTR {
					fmt.Fprintf(stdout, "  %s\n", name)
				}
				fmt.Fprintln(stdout)
			}
		}
	}
}

func parseTypes(t string) []string {
	all := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS"}
	if t == "" {
		return all
	}
	return strings.Split(strings.ToUpper(t), ",")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func printSection(w io.Writer, title string, err error, printContent func()) {
	if err != nil {
		// Don't print "no such host" errors for optional record types, it's just noise
		var dnsErr *net.DNSError
		if isDnsError(err, &dnsErr) && (dnsErr.IsNotFound || strings.Contains(err.Error(), "no such host")) {
			return
		}
		fmt.Fprintf(w, "%s:\n  Error: %v\n\n", title, err)
		return
	}

	// Buffer content to check if it's empty? No, callback handles logic.
	// But we want to print title only if there is content?
	// The current callback approach prints directly.
	// Let's just print title.
	fmt.Fprintf(w, "%s:\n", title)
	printContent()
	fmt.Fprintln(w)
}

func isDnsError(err error, target **net.DNSError) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if dnsErr, ok := opErr.Err.(*net.DNSError); ok {
			*target = dnsErr
			return true
		}
	}
	if dnsErr, ok := err.(*net.DNSError); ok {
		*target = dnsErr
		return true
	}
	return false
}
