package main

import (
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd"
	"github.com/spf13/cobra"
)

func main() {
	boa.CmdT[boa.NoParams]{
		Use:     "tofu",
		Short:   "Tools Of Fantastic Utility",
		Version: appVersion(),
		SubCmds: []*cobra.Command{
			cmd.CatCmd(),
			cmd.FindCmd(),
			cmd.GrepCmd(),
			cmd.Sed2Cmd(),
			cmd.ServeCmd(),
			cmd.UUIDCmd(),
			cmd.PortCmd(),
			cmd.TreeCmd(),
			cmd.WatchCmd(),
			cmd.WhichCmd(),
			cmd.ClipCmd(),
			cmd.IpCmd(),
			cmd.DnsCmd(),
			cmd.HttpCmd(),
		},
	}.Run()
}

func appVersion() string {
	bi, hasBuilInfo := debug.ReadBuildInfo()
	if !hasBuilInfo {
		return "unknown-(no build info)"
	}

	versionString := bi.Main.Version
	if versionString == "" {
		versionString = "unknown-(no version)"
	}

	return versionString
}
