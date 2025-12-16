package tail

import (
	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/k8s/tail/pods"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	return boa.CmdT[boa.NoParams]{
		Use:   "tail",
		Short: "Tail Kubernetes resources",
		SubCmds: []*cobra.Command{
			pods.Cmd(),
		},
	}.ToCobra()
}
