package qr

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

type Params struct {
	Text          string `pos:"true" optional:"true" help:"Text to encode in QR code. If not provided or '-', reads from stdin."`
	RecoveryLevel string `short:"r" optional:"true" help:"Error recovery level (low, medium, high, highest)." default:"medium"`
	Invert        bool   `short:"i" optional:"true" help:"Invert colors (white on black). Default is standard black on white."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "qr",
		Short:       "Render QR codes in the terminal",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if params.Text == "" || params.Text == "-" {
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					data, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
						os.Exit(1)
					}
					params.Text = strings.TrimSpace(string(data))
				}
			}

			if params.Text == "" {
				_ = cmd.Usage()
				os.Exit(1)
			}

			if err := runQr(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runQr(params *Params) error {
	level := qrcode.Medium
	switch strings.ToLower(params.RecoveryLevel) {
	case "low", "l":
		level = qrcode.Low
	case "medium", "m":
		level = qrcode.Medium
	case "high", "h", "q": // 'Q' is technically Quartile (approx 25%), often mapped between M and H, but here let's map to High or keep distinct if lib supports it.
		// go-qrcode supports Low, Medium, High, Highest.
		// Standard QR: L(7%), M(15%), Q(25%), H(30%).
		// go-qrcode: Low, Medium, High, Highest.
		level = qrcode.High
	case "highest":
		level = qrcode.Highest
	}

	qr, err := qrcode.New(params.Text, level)
	if err != nil {
		return fmt.Errorf("generating qr code: %w", err)
	}

	// We render manually to the terminal using ANSI colors or block characters.
	// Standard QR codes are Black modules on White background.
	// Terminals are often Black background.

	// Strategy: Use ANSI background colors.
	// Black Module: \033[40m  \033[0m (Black bg, 2 spaces)
	// White Module: \033[47m  \033[0m (White bg, 2 spaces)

	matrix := qr.Bitmap()

	// By default (no invert), we want standard QR: Black ink on White paper.
	// Black Module (Data) -> ANSI Black BG
	// White Module (Empty) -> ANSI White BG

	var blackStr, whiteStr string

	if params.Invert {
		// "Inverted" (White ink on Black paper)
		// Data (True) = White
		// Empty (False) = Black
		blackStr = "\033[47m  \033[0m" // White block for Data
		whiteStr = "\033[40m  \033[0m" // Black block for Empty
	} else {
		// Standard (Black ink on White paper)
		// Data (True) = Black
		// Empty (False) = White
		blackStr = "\033[40m  \033[0m" // Black block for Data
		whiteStr = "\033[47m  \033[0m" // White block for Empty
	}

	// We assume the bitmap from go-qrcode includes the quiet zone?
	// Checking docs/source... go-qrcode's Bitmap() returns the raw matrix including quiet zone if DisableBorder is false (default).

	for _, row := range matrix {
		for _, col := range row {
			if col {
				fmt.Print(blackStr)
			} else {
				fmt.Print(whiteStr)
			}
		}
		fmt.Println("\033[0m")
	}

	return nil
}
