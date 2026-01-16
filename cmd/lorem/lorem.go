package lorem

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Paragraphs int  `short:"p" help:"Number of paragraphs." default:"0"`
	Sentences  int  `short:"s" help:"Number of sentences." default:"0"`
	Words      int  `short:"w" help:"Number of words." default:"0"`
	StartLorem bool `short:"l" help:"Start with 'Lorem ipsum dolor sit amet'." default:"true"`
}

var words = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "aliquip", "ex", "ea", "commodo",
	"consequat", "duis", "aute", "irure", "in", "reprehenderit", "voluptate",
	"velit", "esse", "cillum", "fugiat", "nulla", "pariatur", "excepteur", "sint",
	"occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum", "at", "vero", "eos",
	"accusamus", "iusto", "odio", "dignissimos", "ducimus", "blanditiis",
	"praesentium", "voluptatum", "deleniti", "atque", "corrupti", "quos", "dolores",
	"quas", "molestias", "excepturi", "obcaecati", "cupiditate", "provident",
	"similique", "mollitia", "animi", "fuga", "harum", "quidem", "rerum", "facilis",
	"expedita", "distinctio", "nam", "libero", "tempore", "cum", "soluta", "nobis",
	"eligendi", "optio", "cumque", "nihil", "impedit", "quo", "minus", "quod",
	"maxime", "placeat", "facere", "possimus", "omnis", "voluptas", "assumenda",
	"repellendus", "temporibus", "autem", "quibusdam", "officiis", "debitis", "aut",
	"reiciendis", "voluptatibus", "maiores", "alias", "perferendis", "doloribus",
	"asperiores", "repellat",
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "lorem",
		Short:       "Generate lorem ipsum text",
		Long:        "Generate placeholder text. Default: 1 paragraph if no options specified.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params)
		},
	}.ToCobra()
}

func Run(params *Params) {
	// Default to 1 paragraph if nothing specified
	if params.Paragraphs == 0 && params.Sentences == 0 && params.Words == 0 {
		params.Paragraphs = 1
	}

	if params.Words > 0 {
		fmt.Println(generateWords(params.Words, params.StartLorem))
	} else if params.Sentences > 0 {
		fmt.Println(generateSentences(params.Sentences, params.StartLorem))
	} else {
		fmt.Println(generateParagraphs(params.Paragraphs, params.StartLorem))
	}
}

func generateWords(count int, startLorem bool) string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		if i == 0 && startLorem {
			result[i] = "Lorem"
		} else if i == 1 && startLorem {
			result[i] = "ipsum"
		} else {
			result[i] = words[rand.Intn(len(words))]
		}
	}
	return strings.Join(result, " ")
}

func generateSentences(count int, startLorem bool) string {
	var sentences []string
	for i := 0; i < count; i++ {
		wordCount := 8 + rand.Intn(10) // 8-17 words per sentence
		useStartLorem := startLorem && i == 0
		sentence := generateWords(wordCount, useStartLorem)
		// Capitalize first letter
		if len(sentence) > 0 {
			sentence = strings.ToUpper(string(sentence[0])) + sentence[1:]
		}
		sentences = append(sentences, sentence+".")
	}
	return strings.Join(sentences, " ")
}

func generateParagraphs(count int, startLorem bool) string {
	var paragraphs []string
	for i := 0; i < count; i++ {
		sentenceCount := 4 + rand.Intn(4) // 4-7 sentences per paragraph
		useStartLorem := startLorem && i == 0
		paragraphs = append(paragraphs, generateSentences(sentenceCount, useStartLorem))
	}
	return strings.Join(paragraphs, "\n\n")
}
