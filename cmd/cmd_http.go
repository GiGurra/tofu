package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type HttpParams struct {
	URL             string   `pos:"true" help:"The URL to request."`
	Method          string   `short:"X" optional:"true" help:"HTTP method to use (GET, POST, PUT, DELETE, etc.). Default is GET." default:"GET"`
	Headers         []string `short:"H" optional:"true" help:"Pass custom header(s) to server."`
	Data            string   `short:"d" optional:"true" help:"HTTP POST data."`
	OutputFile      string   `short:"o" optional:"true" help:"Write to file instead of stdout."`
	FollowRedirects bool     `short:"L" optional:"true" help:"Follow redirects."`
	Verbose         bool     `short:"v" optional:"true" help:"Make the operation more talkative."`
	Insecure        bool     `short:"k" optional:"true" help:"Allow insecure server connections when using SSL."`
}

func HttpCmd() *cobra.Command {
	return boa.CmdT[HttpParams]{
		Use:         "http",
		Short:       "Make HTTP requests (like curl)",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *HttpParams, cmd *cobra.Command, args []string) {
			if params.URL == "" {
				_ = cmd.Usage()
				os.Exit(1)
			}
			// Auto-detect URL scheme if missing
			if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
				params.URL = "http://" + params.URL
			}

			if err := runHttp(params, os.Stdout, os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runHttp(params *HttpParams, stdout, stderr io.Writer) error {
	var body io.Reader
	if params.Data != "" {
		body = strings.NewReader(params.Data)
		// If method is default (GET) and we have data, switch to POST
		if params.Method == "GET" || params.Method == "" {
			params.Method = "POST"
		}
	}

	req, err := http.NewRequest(params.Method, params.URL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	for _, h := range params.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Default User-Agent if not set
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "tofu/http")
	}

	// Configure client
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !params.FollowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	if params.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	if params.Verbose {
		fmt.Fprintf(stderr, "> %s %s %s\n", req.Method, req.URL.Path, req.Proto)
		for name, values := range req.Header {
			for _, value := range values {
				fmt.Fprintf(stderr, "> %s: %s\n", name, value)
			}
		}
		fmt.Fprintln(stderr, ">")
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	if params.Verbose {
		// Print TLS details if available
		if resp.TLS != nil {
			fmt.Fprintf(stderr, "* SSL connection using %s / %s\n", tls.VersionName(resp.TLS.Version), tls.CipherSuiteName(resp.TLS.CipherSuite))
			for i, cert := range resp.TLS.PeerCertificates {
				fmt.Fprintf(stderr, "* Server certificate (%d):\n", i)
				fmt.Fprintf(stderr, "*   subject: %s\n", cert.Subject)
				fmt.Fprintf(stderr, "*   issuer: %s\n", cert.Issuer)
				fmt.Fprintf(stderr, "*   expire date: %s\n", cert.NotAfter.Format(time.RFC822))
			}
		}

		fmt.Fprintf(stderr, "< %s %s\n", resp.Proto, resp.Status)
		for name, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(stderr, "< %s: %s\n", name, value)
			}
		}
		fmt.Fprintln(stderr, "<")
	}

	var out io.Writer = stdout
	if params.OutputFile != "" {
		f, err := os.Create(params.OutputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	return nil
}
