package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type JwtParams struct {
}

func JwtCmd() *cobra.Command {
	return boa.CmdT[JwtParams]{
		Use:   "jwt [token]",
		Short: "Decode and inspect JWT tokens",
		Long: `Decode and inspect JSON Web Tokens (JWT).
The token can be provided as an argument or via standard input.
Displays the decoded Header, Payload (Claims), and the Signature.`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *JwtParams, cmd *cobra.Command, args []string) {
			token := ""
			if len(args) > 0 {
				token = args[0]
			} else {
				// Check if stdin has data
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					data, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
						os.Exit(1)
					}
					token = strings.TrimSpace(string(data))
				}
			}

			if token == "" {
				_ = cmd.Help()
				os.Exit(1)
			}

			if err := runJwt(token); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runJwt(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts (Header.Payload.Signature), found %d", len(parts))
	}

	fmt.Println("Token:")
	fmt.Println(token)
	fmt.Println()

	// Header
	header, err := decodeSegment(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}
	fmt.Println("Header:")
	printJSON(header)
	fmt.Println()

	// Payload
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	fmt.Println("Payload:")
	printJSON(payload)
	fmt.Println()

	// Signature
	fmt.Println("Signature (raw):")
	fmt.Println(parts[2])

	return nil
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

func printJSON(data []byte) {
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		// Fallback if not valid JSON, just print string
		fmt.Println(string(data))
	} else {
		fmt.Println(out.String())
	}
}
