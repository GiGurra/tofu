package uuid

import (
	"fmt"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type Params struct {
	Count     int    `short:"n" help:"Number of UUIDs to generate." default:"1"`
	Version   int    `short:"v" help:"UUID Version (1, 3, 4, 5, 6, 7)." default:"4"`
	Namespace string `short:"s" help:"Namespace for v3/v5 (dns, url, oid, x500, or UUID string)." default:""`
	Name      string `short:"d" help:"Data/Name for v3/v5 generation." default:""`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "uuid",
		Short:       "Generate UUIDs",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				fmt.Fprintf(os.Stderr, "uuid: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	for i := 0; i < params.Count; i++ {
		var u uuid.UUID
		var err error

		switch params.Version {
		case 1:
			u, err = uuid.NewUUID()
		case 3:
			ns, nsErr := ParseNamespace(params.Namespace)
			if nsErr != nil {
				return nsErr
			}
			if params.Name == "" {
				return fmt.Errorf("v3 requires --name/-d")
			}
			u = uuid.NewMD5(ns, []byte(params.Name))
		case 4:
			u, err = uuid.NewRandom()
		case 5:
			ns, nsErr := ParseNamespace(params.Namespace)
			if nsErr != nil {
				return nsErr
			}
			if params.Name == "" {
				return fmt.Errorf("v5 requires --name/-d")
			}
			u = uuid.NewSHA1(ns, []byte(params.Name))
		case 6:
			u, err = uuid.NewV6()
		case 7:
			u, err = uuid.NewV7()
		default:
			return fmt.Errorf("unsupported UUID version: %d", params.Version)
		}

		if err != nil {
			return fmt.Errorf("failed to generate UUID: %w", err)
		}

		fmt.Println(u.String())
	}
	return nil
}

func ParseNamespace(ns string) (uuid.UUID, error) {
	switch strings.ToLower(ns) {
	case "dns":
		return uuid.NameSpaceDNS, nil
	case "url":
		return uuid.NameSpaceURL, nil
	case "oid":
		return uuid.NameSpaceOID, nil
	case "x500":
		return uuid.NameSpaceX500, nil
	case "":
		return uuid.Nil, fmt.Errorf("v3/v5 requires --namespace/-s")
	default:
		// Try to parse as UUID
		return uuid.Parse(ns)
	}
}
