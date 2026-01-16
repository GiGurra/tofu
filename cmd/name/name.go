package name

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Count  int    `short:"n" help:"Number of names to generate." default:"1"`
	Style  string `short:"s" help:"Style: operation, project, variable, animal." default:"operation"`
}

var adjectives = []string{
	"thundering", "silent", "golden", "crimson", "azure", "phantom", "cosmic",
	"electric", "frozen", "burning", "shadow", "crystal", "iron", "velvet",
	"emerald", "obsidian", "silver", "ancient", "savage", "noble", "midnight",
	"stellar", "quantum", "rapid", "stealth", "primal", "cyber", "turbo",
	"mega", "ultra", "hyper", "super", "infinite", "eternal", "atomic",
	"dynamic", "binary", "digital", "viral", "neural", "spectral", "lunar",
}

var nouns = []string{
	"gopher", "phoenix", "dragon", "falcon", "panther", "wolf", "tiger",
	"eagle", "hawk", "cobra", "viper", "shark", "thunder", "storm", "blaze",
	"frost", "shadow", "nova", "nebula", "pulsar", "quasar", "horizon",
	"vertex", "matrix", "cipher", "nexus", "apex", "zenith", "omega",
	"prime", "core", "flux", "spark", "pulse", "wave", "surge", "forge",
	"anvil", "hammer", "shield", "sword", "arrow", "spear", "blade",
}

var techPrefixes = []string{
	"go", "rust", "node", "react", "vue", "next", "fast", "quick", "swift",
	"smart", "auto", "sync", "async", "meta", "micro", "nano", "poly", "mono",
}

var techSuffixes = []string{
	"ify", "ly", "io", "hub", "lab", "kit", "box", "base", "flow", "sync",
	"stack", "cloud", "forge", "craft", "works", "wave", "pulse", "core",
}

var animals = []string{
	"aardvark", "badger", "capybara", "dingo", "elephant", "flamingo",
	"giraffe", "hedgehog", "iguana", "jellyfish", "koala", "lemur",
	"meerkat", "narwhal", "octopus", "pangolin", "quokka", "raccoon",
	"sloth", "tapir", "uakari", "vulture", "wombat", "xerus", "yak", "zebra",
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "name",
		Short:       "Generate random project/operation names",
		Long:        "Generate random names for projects, operations, variables, or just fun.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	for i := 0; i < params.Count; i++ {
		switch strings.ToLower(params.Style) {
		case "operation", "op":
			fmt.Printf("Operation %s %s\n", capitalize(adjectives[rand.Intn(len(adjectives))]), capitalize(nouns[rand.Intn(len(nouns))]))
		case "project", "proj":
			fmt.Println(techPrefixes[rand.Intn(len(techPrefixes))] + techSuffixes[rand.Intn(len(techSuffixes))])
		case "variable", "var":
			fmt.Println(strings.ToLower(adjectives[rand.Intn(len(adjectives))]) + capitalize(nouns[rand.Intn(len(nouns))]))
		case "animal":
			fmt.Printf("%s %s\n", capitalize(adjectives[rand.Intn(len(adjectives))]), animals[rand.Intn(len(animals))])
		default:
			fmt.Printf("Operation %s %s\n", capitalize(adjectives[rand.Intn(len(adjectives))]), capitalize(nouns[rand.Intn(len(nouns))]))
		}
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}
