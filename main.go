package main

import (
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/archive"
	"github.com/gigurra/tofu/cmd/base64"
	"github.com/gigurra/tofu/cmd/bird"
	"github.com/gigurra/tofu/cmd/cat"
	"github.com/gigurra/tofu/cmd/clip"
	"github.com/gigurra/tofu/cmd/count"
	"github.com/gigurra/tofu/cmd/cp"
	"github.com/gigurra/tofu/cmd/cron"
	"github.com/gigurra/tofu/cmd/df"
	"github.com/gigurra/tofu/cmd/diff"
	"github.com/gigurra/tofu/cmd/dns"
	"github.com/gigurra/tofu/cmd/du"
	"github.com/gigurra/tofu/cmd/env"
	"github.com/gigurra/tofu/cmd/find"
	"github.com/gigurra/tofu/cmd/free"
	"github.com/gigurra/tofu/cmd/gh"
	"github.com/gigurra/tofu/cmd/grep"
	"github.com/gigurra/tofu/cmd/hash"
	"github.com/gigurra/tofu/cmd/head"
	"github.com/gigurra/tofu/cmd/http"
	"github.com/gigurra/tofu/cmd/ip"
	"github.com/gigurra/tofu/cmd/jwt"
	"github.com/gigurra/tofu/cmd/k8s"
	"github.com/gigurra/tofu/cmd/ls"
	"github.com/gigurra/tofu/cmd/mkdir"
	"github.com/gigurra/tofu/cmd/mv"
	"github.com/gigurra/tofu/cmd/nc"
	"github.com/gigurra/tofu/cmd/ping"
	"github.com/gigurra/tofu/cmd/port"
	"github.com/gigurra/tofu/cmd/ps"
	"github.com/gigurra/tofu/cmd/qr"
	"github.com/gigurra/tofu/cmd/rand"
	"github.com/gigurra/tofu/cmd/reverse"
	"github.com/gigurra/tofu/cmd/rm"
	"github.com/gigurra/tofu/cmd/screensaver"
	"github.com/gigurra/tofu/cmd/sed2"
	"github.com/gigurra/tofu/cmd/serve"
	"github.com/gigurra/tofu/cmd/tail"
	"github.com/gigurra/tofu/cmd/tee"
	"github.com/gigurra/tofu/cmd/time"
	"github.com/gigurra/tofu/cmd/touch"
	"github.com/gigurra/tofu/cmd/tree"
	"github.com/gigurra/tofu/cmd/uuid"
	"github.com/gigurra/tofu/cmd/watch"
	"github.com/gigurra/tofu/cmd/wget"
	"github.com/gigurra/tofu/cmd/which"
	"github.com/spf13/cobra"
)

func main() {
	boa.CmdT[boa.NoParams]{
		Use:     "tofu",
		Short:   "Tools Of Fantastic Utility",
		Version: appVersion(),
		SubCmds: []*cobra.Command{
			cat.Cmd(),
			find.Cmd(),
			grep.Cmd(),
			sed2.Cmd(),
			serve.Cmd(),
			uuid.Cmd(),
			port.Cmd(),
			tree.Cmd(),
			watch.Cmd(),
			which.Cmd(),
			clip.Cmd(),
			ip.Cmd(),
			dns.Cmd(),
			http.Cmd(),
			nc.Cmd(),
			qr.Cmd(),
			base64.Cmd(),
			jwt.Cmd(),
			hash.Cmd(),
			free.Cmd(),
			ps.Cmd(),
			time.Cmd(),
			rand.Cmd(),
			reverse.Cmd(),
			env.Cmd(),
			cron.Cmd(),
			count.Cmd(),
			diff.Cmd(),
			tail.Cmd(),
			tee.Cmd(),
			touch.Cmd(),
			head.Cmd(),
			archive.Cmd(),
			wget.Cmd(),
			screensaver.Cmd(),
			bird.Cmd(),
			k8s.Cmd(),
			gh.Cmd(),
			du.Cmd(),
			df.Cmd(),
			ls.Cmd(),
			ls.LlCmd(),
			ls.LaCmd(),
			mkdir.Cmd(),
			mv.Cmd(),
			cp.Cmd(),
			rm.Cmd(),
			ping.Cmd(),
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
