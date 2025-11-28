package dns

import (
	"context"
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

	fmt.Fprintf(stdout, "Server:  %s\n", serverName)
	fmt.Fprintf(stdout, "Address: %s\n\n", params.Hostname)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	typesToQuery := parseTypes(params.Types)

	for _, recordType := range typesToQuery {
		switch recordType {
		case "A":
			ips, err := resolver.LookupIPAddr(ctx, params.Hostname)
			printSection(stdout, "A Records", err, func() {
				found := false
				for _, ip := range ips {
					if ip.IP.To4() != nil {
						fmt.Fprintf(stdout, "  %s\n", ip.String())
						found = true
					}
				}
				if !found && err == nil {
					fmt.Fprintln(stdout, "  (None)")
				}
			})

		case "AAAA":
			ips, err := resolver.LookupIPAddr(ctx, params.Hostname)
			printSection(stdout, "AAAA Records", err, func() {
				found := false
				for _, ip := range ips {
					if ip.IP.To4() == nil {
						fmt.Fprintf(stdout, "  %s\n", ip.String())
						found = true
					}
				}
				if !found && err == nil {
					fmt.Fprintln(stdout, "  (None)")
				}
			})

		case "CNAME":
			cname, err := resolver.LookupCNAME(ctx, params.Hostname)
			// LookupCNAME often returns error if no CNAME (it returns the name itself if it's canonical).
			// Standard lib behavior: "The returned CNAME is followed to its end."
			printSection(stdout, "CNAME", err, func() {
				if cname != "" && cname != params.Hostname && cname != params.Hostname+"." {
					fmt.Fprintf(stdout, "  %s\n", cname)
				} else if err == nil {
					fmt.Fprintf(stdout, "  (None)\n")
				}
			})

		case "MX":
			mxs, err := resolver.LookupMX(ctx, params.Hostname)
			printSection(stdout, "MX Records", err, func() {
				// Sort by pref
				sort.Slice(mxs, func(i, j int) bool { return mxs[i].Pref < mxs[j].Pref })
				for _, mx := range mxs {
					fmt.Fprintf(stdout, "  %d %s\n", mx.Pref, mx.Host)
				}
			})

		case "TXT":
			txts, err := resolver.LookupTXT(ctx, params.Hostname)
			printSection(stdout, "TXT Records", err, func() {
				for _, txt := range txts {
					fmt.Fprintf(stdout, "  %s\n", txt)
				}
			})

		case "NS":
			nss, err := resolver.LookupNS(ctx, params.Hostname)
			printSection(stdout, "NS Records", err, func() {
				for _, ns := range nss {
					fmt.Fprintf(stdout, "  %s\n", ns.Host)
				}
			})

		case "PTR":
			// Usually for IPs, but user might ask for it on a hostname
			names, err := resolver.LookupAddr(ctx, params.Hostname)
			printSection(stdout, "PTR Records", err, func() {
				for _, name := range names {
					fmt.Fprintf(stdout, "  %s\n", name)
				}
			})
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
