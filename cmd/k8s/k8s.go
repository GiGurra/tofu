package k8s

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/k8s/tail"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "k8s",
		Short: "Kubernetes utilities",
		SubCmds: []*cobra.Command{
			tail.Cmd(),
		},
	}.ToCobra()
}
