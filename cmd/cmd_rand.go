package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
	"go.1password.io/spg"
)

type RandParams struct {
	Type       string `short:"t" help:"Type of random data (str, int, hex, base64, password, phrase)." default:"str"`
	Length     int    `short:"l" help:"Length (chars for str/password/hex/base64, words for phrase)." default:"16"`
	Min        int64  `help:"Minimum value for integer generation." default:"0"`
	Max        int64  `help:"Maximum value for integer generation." default:"100"`
	Charset    string `short:"c" help:"Custom character set for string generation." default:""`
	Count      int    `short:"n" help:"Number of items to generate." default:"1"`
	Separator  string `help:"Separator for phrases." default:" "`
	Capitalize string `help:"Capitalization for phrases (none, first, all, random, one)." default:"none"`
}

func RandCmd() *cobra.Command {
	return boa.CmdT[RandParams]{
		Use:         "rand",
		Short:       "Generate random data",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *RandParams, cmd *cobra.Command, args []string) {
			if err := runRand(params); err != nil {
				fmt.Fprintf(os.Stderr, "rand: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runRand(params *RandParams) error {
	for i := 0; i < params.Count; i++ {
		val, err := generateRandom(params)
		if err != nil {
			return err
		}
		fmt.Println(val)
	}
	return nil
}

func generateRandom(params *RandParams) (string, error) {
	switch params.Type {
	case "int":
		if params.Min > params.Max {
			return "", fmt.Errorf("min cannot be greater than max")
		}
		diff := new(big.Int).Sub(big.NewInt(params.Max), big.NewInt(params.Min))
		diff.Add(diff, big.NewInt(1)) // inclusive max
		if diff.Sign() <= 0 {
			// should be handled by check above but for big int safety
			return "", fmt.Errorf("invalid range")
		}
		n, err := rand.Int(rand.Reader, diff)
		if err != nil {
			return "", err
		}
		return new(big.Int).Add(n, big.NewInt(params.Min)).String(), nil

	case "str":
		charset := params.Charset
		if charset == "" {
			charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		}
		return randomString(params.Length, charset)

	case "hex":
		b := make([]byte, params.Length)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(b), nil

	case "base64":
		b := make([]byte, params.Length)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil

	case "password":
		r := spg.NewCharRecipe(params.Length)
		r.Allow = spg.Letters | spg.Digits | spg.Symbols
		// We require at least one of each to ensure strength, 
		// but spg defaults might differ. Let's enforce strong defaults.
		r.Require = spg.Letters | spg.Digits | spg.Symbols
		
		pwd, err := r.Generate()
		if err != nil {
			return "", err
		}
		return pwd.String(), nil

	case "phrase":
		// Use AgileWords by default
		wl, err := spg.NewWordList(spg.AgileWords)
		if err != nil {
			return "", err
		}
		r := spg.NewWLRecipe(params.Length, wl)
		r.SeparatorChar = params.Separator
		
		switch params.Capitalize {
		case "first":
			r.Capitalize = spg.CSFirst
		case "all":
			r.Capitalize = spg.CSAll
		case "random":
			r.Capitalize = spg.CSRandom
		case "one":
			r.Capitalize = spg.CSOne
		default:
			r.Capitalize = spg.CSNone
		}

		pwd, err := r.Generate()
		if err != nil {
			return "", err
		}
		return pwd.String(), nil

	default:
		return "", fmt.Errorf("unknown type: %s", params.Type)
	}
}

func randomString(length int, charset string) (string, error) {
	b := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}