package figlet

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Text []string `pos:"true" optional:"true" help:"Text to render. If none provided, reads from stdin."`
	Font string   `short:"f" help:"Font: standard, small, mini, block." default:"standard"`
}

// Simple block-style font
var standardFont = map[rune][5]string{
	'A': {"  █  ", " █ █ ", "█████", "█   █", "█   █"},
	'B': {"████ ", "█   █", "████ ", "█   █", "████ "},
	'C': {" ████", "█    ", "█    ", "█    ", " ████"},
	'D': {"████ ", "█   █", "█   █", "█   █", "████ "},
	'E': {"█████", "█    ", "████ ", "█    ", "█████"},
	'F': {"█████", "█    ", "████ ", "█    ", "█    "},
	'G': {" ████", "█    ", "█  ██", "█   █", " ████"},
	'H': {"█   █", "█   █", "█████", "█   █", "█   █"},
	'I': {"█████", "  █  ", "  █  ", "  █  ", "█████"},
	'J': {"█████", "   █ ", "   █ ", "█  █ ", " ██  "},
	'K': {"█   █", "█  █ ", "███  ", "█  █ ", "█   █"},
	'L': {"█    ", "█    ", "█    ", "█    ", "█████"},
	'M': {"█   █", "██ ██", "█ █ █", "█   █", "█   █"},
	'N': {"█   █", "██  █", "█ █ █", "█  ██", "█   █"},
	'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
	'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
	'Q': {" ███ ", "█   █", "█ █ █", "█  █ ", " ██ █"},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'S': {" ████", "█    ", " ███ ", "    █", "████ "},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'U': {"█   █", "█   █", "█   █", "█   █", " ███ "},
	'V': {"█   █", "█   █", "█   █", " █ █ ", "  █  "},
	'W': {"█   █", "█   █", "█ █ █", "██ ██", "█   █"},
	'X': {"█   █", " █ █ ", "  █  ", " █ █ ", "█   █"},
	'Y': {"█   █", " █ █ ", "  █  ", "  █  ", "  █  "},
	'Z': {"█████", "   █ ", "  █  ", " █   ", "█████"},
	'0': {" ███ ", "█  ██", "█ █ █", "██  █", " ███ "},
	'1': {"  █  ", " ██  ", "  █  ", "  █  ", "█████"},
	'2': {" ███ ", "█   █", "  ██ ", " █   ", "█████"},
	'3': {"████ ", "    █", " ███ ", "    █", "████ "},
	'4': {"█   █", "█   █", "█████", "    █", "    █"},
	'5': {"█████", "█    ", "████ ", "    █", "████ "},
	'6': {" ███ ", "█    ", "████ ", "█   █", " ███ "},
	'7': {"█████", "    █", "   █ ", "  █  ", "  █  "},
	'8': {" ███ ", "█   █", " ███ ", "█   █", " ███ "},
	'9': {" ███ ", "█   █", " ████", "    █", " ███ "},
	' ': {"     ", "     ", "     ", "     ", "     "},
	'!': {"  █  ", "  █  ", "  █  ", "     ", "  █  "},
	'?': {" ███ ", "█   █", "  ██ ", "     ", "  █  "},
	'.': {"     ", "     ", "     ", "     ", "  █  "},
	',': {"     ", "     ", "     ", "  █  ", " █   "},
	'-': {"     ", "     ", "█████", "     ", "     "},
	'_': {"     ", "     ", "     ", "     ", "█████"},
	':': {"     ", "  █  ", "     ", "  █  ", "     "},
}

var smallFont = map[rune][3]string{
	'A': {"▄█▄", "█▀█", "▀ ▀"},
	'B': {"██▄", "█▄█", "██▀"},
	'C': {"▄█▀", "█  ", "▀█▄"},
	'D': {"██▄", "█ █", "██▀"},
	'E': {"██▀", "█▄ ", "██▄"},
	'F': {"██▀", "█▄ ", "█  "},
	'G': {"▄█▀", "█ █", "▀█▄"},
	'H': {"█ █", "███", "█ █"},
	'I': {"███", " █ ", "███"},
	'J': {"▀▀█", "  █", "▀█▀"},
	'K': {"█▄▀", "██ ", "█ █"},
	'L': {"█  ", "█  ", "███"},
	'M': {"█▄█", "█▀█", "█ █"},
	'N': {"█▀█", "█▀█", "█ █"},
	'O': {"▄█▄", "█ █", "▀█▀"},
	'P': {"██▄", "█▀▀", "█  "},
	'Q': {"▄█▄", "█ █", "▀█▄"},
	'R': {"██▄", "██▀", "█ █"},
	'S': {"▄█▀", "▀█▄", "▀█▄"},
	'T': {"███", " █ ", " █ "},
	'U': {"█ █", "█ █", "▀█▀"},
	'V': {"█ █", "█ █", " ▀ "},
	'W': {"█ █", "█▀█", "▀▀▀"},
	'X': {"█ █", " ▀ ", "█ █"},
	'Y': {"█ █", " █ ", " █ "},
	'Z': {"▀▀█", " █ ", "█▀▀"},
	' ': {"   ", "   ", "   "},
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "figlet",
		Short:       "ASCII art text banners",
		Long:        "Render text as large ASCII art banners.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	var text string

	if len(params.Text) > 0 {
		text = strings.Join(params.Text, " ")
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text = scanner.Text()
		}
	}

	if text == "" {
		text = "HELLO"
	}

	text = strings.ToUpper(text)

	switch params.Font {
	case "small":
		renderSmall(text)
	case "mini":
		renderMini(text)
	case "block":
		renderBlock(text)
	default:
		renderStandard(text)
	}
}

func renderStandard(text string) {
	for row := 0; row < 5; row++ {
		for _, char := range text {
			if glyph, ok := standardFont[char]; ok {
				fmt.Print(glyph[row])
			} else {
				fmt.Print("     ")
			}
		}
		fmt.Println()
	}
}

func renderSmall(text string) {
	for row := 0; row < 3; row++ {
		for _, char := range text {
			if glyph, ok := smallFont[char]; ok {
				fmt.Print(glyph[row])
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Println()
	}
}

func renderMini(text string) {
	// Use unicode block characters for tiny text
	fmt.Println(text)
}

func renderBlock(text string) {
	// Each character becomes a 3x3 block
	for row := 0; row < 3; row++ {
		for _, char := range text {
			if char == ' ' {
				fmt.Print("   ")
			} else {
				fmt.Print("███")
			}
			fmt.Print(" ")
		}
		fmt.Println()
	}
}
