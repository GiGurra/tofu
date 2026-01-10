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
	"sync"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Hostname string   `pos:"true" help:"Hostname to lookup"`
	Server   string   `short:"s" help:"DNS server to use. Use 'os' for OS resolver, or IP address (e.g. 8.8.8.8)" default:"os" alts:"os,8.8.8.8,1.1.1.1" strict:"false"`
	Types    []string `short:"t" help:"Record types to query. Use 'all' for all types. Default: A,AAAA,CNAME" default:"A,AAAA,CNAME" alts:"A,AAAA,CNAME,MX,TXT,NS,PTR,all"`
	Timeout  int      `long:"timeout" help:"Timeout in seconds for DNS queries" default:"2"`
	Json     bool     `short:"j" help:"Output in JSON format."`
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

	useOS := strings.ToLower(params.Server) == "os"

	if useOS {
		resolver = net.DefaultResolver
		serverName = "OS"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(params.Timeout)*time.Second)
	defer cancel()

	typesToQuery := parseTypes(params.Types)

	// Track errors for non-JSON output
	type recordError struct {
		recordType string
		err        error
	}
	var errors []recordError
	var errorsMu sync.Mutex

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, recordType := range typesToQuery {
		switch recordType {
		case "A":
			wg.Add(1)
			go func() {
				defer wg.Done()
				var ips []net.IP
				var err error
				if useOS {
					ips, err = lookupHostCgo(params.Hostname)
				} else {
					var ipAddrs []net.IPAddr
					ipAddrs, err = resolver.LookupIPAddr(ctx, params.Hostname)
					for _, ip := range ipAddrs {
						ips = append(ips, ip.IP)
					}
				}
				mu.Lock()
				if err == nil {
					for _, ip := range ips {
						if ip.To4() != nil {
							output.A = append(output.A, ip.String())
						}
					}
				} else {
					errorsMu.Lock()
					errors = append(errors, recordError{"A Records", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()

		case "AAAA":
			wg.Add(1)
			go func() {
				defer wg.Done()
				var ips []net.IP
				var err error
				if useOS {
					ips, err = lookupHostCgo(params.Hostname)
				} else {
					var ipAddrs []net.IPAddr
					ipAddrs, err = resolver.LookupIPAddr(ctx, params.Hostname)
					for _, ip := range ipAddrs {
						ips = append(ips, ip.IP)
					}
				}
				mu.Lock()
				if err == nil {
					for _, ip := range ips {
						if ip.To4() == nil {
							output.AAAA = append(output.AAAA, ip.String())
						}
					}
				} else {
					errorsMu.Lock()
					errors = append(errors, recordError{"AAAA Records", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()

		case "CNAME":
			wg.Add(1)
			go func() {
				defer wg.Done()
				var cname string
				var err error
				if useOS {
					cname, err = lookupCNAMECgo(params.Hostname)
				} else {
					cname, err = resolver.LookupCNAME(ctx, params.Hostname)
				}
				mu.Lock()
				if err == nil && cname != "" && cname != params.Hostname && cname != params.Hostname+"." {
					output.CNAME = cname
				} else if err != nil {
					errorsMu.Lock()
					errors = append(errors, recordError{"CNAME", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()

		case "MX":
			wg.Add(1)
			go func() {
				defer wg.Done()
				if useOS {
					mxs, err := lookupMXCgo(params.Hostname)
					mu.Lock()
					if err == nil {
						sort.Slice(mxs, func(i, j int) bool { return mxs[i].Pref < mxs[j].Pref })
						for _, mx := range mxs {
							output.MX = append(output.MX, MXRecord{Pref: mx.Pref, Host: mx.Host})
						}
					} else {
						errorsMu.Lock()
						errors = append(errors, recordError{"MX Records", err})
						errorsMu.Unlock()
					}
					mu.Unlock()
				} else {
					mxs, err := resolver.LookupMX(ctx, params.Hostname)
					mu.Lock()
					if err == nil {
						sort.Slice(mxs, func(i, j int) bool { return mxs[i].Pref < mxs[j].Pref })
						for _, mx := range mxs {
							output.MX = append(output.MX, MXRecord{Pref: mx.Pref, Host: mx.Host})
						}
					} else {
						errorsMu.Lock()
						errors = append(errors, recordError{"MX Records", err})
						errorsMu.Unlock()
					}
					mu.Unlock()
				}
			}()

		case "TXT":
			wg.Add(1)
			go func() {
				defer wg.Done()
				var txts []string
				var err error
				if useOS {
					txts, err = lookupTXTCgo(params.Hostname)
				} else {
					txts, err = resolver.LookupTXT(ctx, params.Hostname)
				}
				mu.Lock()
				if err == nil {
					output.TXT = txts
				} else {
					errorsMu.Lock()
					errors = append(errors, recordError{"TXT Records", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()

		case "NS":
			wg.Add(1)
			go func() {
				defer wg.Done()
				var nss []string
				var err error
				if useOS {
					nss, err = lookupNSCgo(params.Hostname)
				} else {
					netNss, netErr := resolver.LookupNS(ctx, params.Hostname)
					err = netErr
					if netErr == nil {
						for _, ns := range netNss {
							nss = append(nss, ns.Host)
						}
					}
				}
				mu.Lock()
				if err == nil {
					output.NS = nss
				} else {
					errorsMu.Lock()
					errors = append(errors, recordError{"NS Records", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()

		case "PTR":
			wg.Add(1)
			go func() {
				defer wg.Done()
				names, err := resolver.LookupAddr(ctx, params.Hostname)
				mu.Lock()
				if err == nil {
					output.PTR = names
				} else {
					errorsMu.Lock()
					errors = append(errors, recordError{"PTR Records", err})
					errorsMu.Unlock()
				}
				mu.Unlock()
			}()
		}
	}

	wg.Wait()

	// Print errors for non-JSON output
	if !params.Json {
		for _, e := range errors {
			printSection(stdout, e.recordType, e.err, func() {})
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

func parseTypes(types []string) []string {
	all := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "PTR"}
	if len(types) == 0 {
		return []string{"A", "AAAA", "CNAME"}
	}
	// Check for "all" option
	for _, t := range types {
		if strings.ToUpper(t) == "ALL" {
			return all
		}
	}
	// Normalize to uppercase
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = strings.ToUpper(t)
	}
	return result
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
