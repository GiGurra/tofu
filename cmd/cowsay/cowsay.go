package cowsay

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
	Message []string `pos:"true" optional:"true" help:"Message to say. If none provided, reads from stdin."`
	Animal  string   `short:"a" help:"Animal: cow, tux, tofu, gopher, cat, ghost." default:"cow"`
	Think   bool     `short:"t" help:"Think instead of say (use thought bubble)." default:"false"`
}

var animals = map[string]string{
	"cow": `        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||`,
	"tux": `       \
        \
            .--.
           |o_o |
           |:_/ |
          //   \ \
         (|     | )
        /'\_   _/` + "`" + `\
        \___)=(___/`,
	"tofu": `       \
        \  ___________
         | |  TOFU  | |
         | |________| |
          \__________/
           |        |
           |________|`,
	"gopher": `       \
        \
          ʕ◔ϖ◔ʔ
         /    \
        | ^  ^ |
         \    /
          ~~~~`,
	"cat": `       \
        \    /\_/\
         \  ( o.o )
            > ^ <
           /|   |\
          (_|   |_)`,
	"ghost": `       \
        \   ___
          /    \
         | ^  ^ |
         |  __  |
          \    /
           ^^^^`,
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "cowsay",
		Short:       "Make an ASCII animal say something",
		Long:        "Generate an ASCII picture of an animal saying something. Classic!",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	var message string

	if len(params.Message) > 0 {
		message = strings.Join(params.Message, " ")
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		message = strings.Join(lines, " ")
	}

	if message == "" {
		message = "Moo!"
	}

	// Wrap message
	maxWidth := 40
	wrapped := wrapText(message, maxWidth)

	// Build speech bubble
	bubbleChar := "o"
	if !params.Think {
		bubbleChar = "\\"
	}

	// Get bubble width
	bubbleWidth := 0
	for _, line := range wrapped {
		if len(line) > bubbleWidth {
			bubbleWidth = len(line)
		}
	}

	// Draw bubble
	fmt.Println(" " + strings.Repeat("_", bubbleWidth+2))
	for i, line := range wrapped {
		padding := strings.Repeat(" ", bubbleWidth-len(line))
		if len(wrapped) == 1 {
			fmt.Printf("< %s%s >\n", line, padding)
		} else if i == 0 {
			fmt.Printf("/ %s%s \\\n", line, padding)
		} else if i == len(wrapped)-1 {
			fmt.Printf("\\ %s%s /\n", line, padding)
		} else {
			fmt.Printf("| %s%s |\n", line, padding)
		}
	}
	fmt.Println(" " + strings.Repeat("-", bubbleWidth+2))

	// Draw animal
	animal, ok := animals[strings.ToLower(params.Animal)]
	if !ok {
		animal = animals["cow"]
	}

	// Replace \ with o for think mode
	if params.Think {
		animal = strings.ReplaceAll(animal, "\\", bubbleChar)
	}

	fmt.Println(animal)
}

func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
