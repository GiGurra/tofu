package typing

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Words int `short:"w" help:"Number of words to type." default:"25"`
}

var wordList = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
	"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
	"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
	"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
	"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
	"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
	"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
	"even", "new", "want", "because", "any", "these", "give", "day", "most", "us",
	"code", "function", "variable", "class", "method", "return", "import", "export",
	"const", "let", "var", "async", "await", "promise", "error", "debug", "test",
	"build", "deploy", "server", "client", "database", "query", "cache", "memory",
	"string", "number", "boolean", "array", "object", "null", "undefined", "type",
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "typing",
		Short:       "Typing speed test",
		Long:        "Test your typing speed. Type the displayed words as fast as you can!",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	// Generate random words
	words := make([]string, params.Words)
	for i := range words {
		words[i] = wordList[rand.Intn(len(wordList))]
	}
	text := strings.Join(words, " ")

	// Clear screen
	fmt.Print("\033[2J\033[H")

	fmt.Println("⌨️  TYPING SPEED TEST")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Type the following text. Press Enter when done.")
	fmt.Println()

	// Print text to type (with word wrapping)
	printWrapped(text, 60)

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Print("> ")

	// Start timer when user starts typing
	reader := bufio.NewReader(os.Stdin)
	startTime := time.Now()
	input, _ := reader.ReadString('\n')
	endTime := time.Now()

	input = strings.TrimSpace(input)
	duration := endTime.Sub(startTime)

	// Calculate statistics
	inputWords := strings.Fields(input)
	targetWords := strings.Fields(text)

	// Count correct characters
	correctChars := 0
	totalChars := len(text)
	minLen := len(input)
	if len(text) < minLen {
		minLen = len(text)
	}
	for i := 0; i < minLen; i++ {
		if input[i] == text[i] {
			correctChars++
		}
	}

	// Count correct words
	correctWords := 0
	for i := 0; i < len(inputWords) && i < len(targetWords); i++ {
		if inputWords[i] == targetWords[i] {
			correctWords++
		}
	}

	// Calculate WPM (standard: 5 characters = 1 word)
	minutes := duration.Minutes()
	grossWPM := float64(len(input)) / 5.0 / minutes
	accuracy := float64(correctChars) / float64(totalChars) * 100
	netWPM := grossWPM * (accuracy / 100)

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 RESULTS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("⏱️  Time:      %.1f seconds\n", duration.Seconds())
	fmt.Printf("📝 Gross WPM: %.0f\n", grossWPM)
	fmt.Printf("✅ Net WPM:   %.0f\n", netWPM)
	fmt.Printf("🎯 Accuracy:  %.1f%%\n", accuracy)
	fmt.Printf("📖 Words:     %d/%d correct\n", correctWords, len(targetWords))
	fmt.Println()

	// Rating
	switch {
	case netWPM >= 80:
		fmt.Println("🏆 LEGENDARY! Are you a court stenographer?")
	case netWPM >= 60:
		fmt.Println("🥇 Excellent! You're a typing wizard!")
	case netWPM >= 40:
		fmt.Println("🥈 Good job! Above average typing speed.")
	case netWPM >= 25:
		fmt.Println("🥉 Not bad! Keep practicing.")
	default:
		fmt.Println("💪 Keep at it! Practice makes perfect.")
	}
}

func printWrapped(text string, width int) {
	words := strings.Fields(text)
	line := ""
	for _, word := range words {
		if len(line)+len(word)+1 > width {
			fmt.Println(line)
			line = word
		} else if line == "" {
			line = word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		fmt.Println(line)
	}
}
