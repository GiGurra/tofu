package weather

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Location string `pos:"true" optional:"true" help:"Location (city name, airport code, or coordinates)." default:""`
	Format   string `short:"f" help:"Format: full, short, oneline." default:"short"`
	Units    string `short:"u" help:"Units: metric (m), imperial (u)." default:"m"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "weather",
		Short:       "Display ASCII weather",
		Long:        "Fetch and display weather using wttr.in. Shows ASCII art weather for any location.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				fmt.Fprintf(os.Stderr, "weather: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	// Build wttr.in URL
	baseURL := "https://wttr.in/"

	location := params.Location
	if location != "" {
		location = url.PathEscape(location)
	}

	// Build format string
	var format string
	switch params.Format {
	case "full":
		format = ""
	case "oneline":
		format = "?format=3"
	default: // short
		format = "?0"
	}

	// Add units
	if format == "" {
		format = "?" + params.Units
	} else if format == "?0" {
		format += "&" + params.Units
	}

	reqURL := baseURL + location + format

	// Make request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// wttr.in uses User-Agent to detect terminal vs browser
	req.Header.Set("User-Agent", "curl/7.68.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wttr.in returned status %d", resp.StatusCode)
	}

	// Stream output
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
