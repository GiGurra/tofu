package main

import (
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/archive"
	"github.com/gigurra/tofu/cmd/base64"
	"github.com/gigurra/tofu/cmd/bird"
	"github.com/gigurra/tofu/cmd/blame"
	"github.com/gigurra/tofu/cmd/busy"
	"github.com/gigurra/tofu/cmd/calendar"
	"github.com/gigurra/tofu/cmd/cat"
	"github.com/gigurra/tofu/cmd/clip"
	"github.com/gigurra/tofu/cmd/clock"
	"github.com/gigurra/tofu/cmd/coin"
	"github.com/gigurra/tofu/cmd/count"
	"github.com/gigurra/tofu/cmd/cowsay"
	"github.com/gigurra/tofu/cmd/cp"
	"github.com/gigurra/tofu/cmd/cron"
	"github.com/gigurra/tofu/cmd/crypt"
	"github.com/gigurra/tofu/cmd/df"
	"github.com/gigurra/tofu/cmd/dice"
	"github.com/gigurra/tofu/cmd/diff"
	"github.com/gigurra/tofu/cmd/dns"
	"github.com/gigurra/tofu/cmd/du"
	"github.com/gigurra/tofu/cmd/env"
	"github.com/gigurra/tofu/cmd/excuse"
	"github.com/gigurra/tofu/cmd/figlet"
	"github.com/gigurra/tofu/cmd/find"
	"github.com/gigurra/tofu/cmd/flip"
	"github.com/gigurra/tofu/cmd/fortune"
	"github.com/gigurra/tofu/cmd/free"
	"github.com/gigurra/tofu/cmd/gh"
	"github.com/gigurra/tofu/cmd/git"
	"github.com/gigurra/tofu/cmd/grep"
	"github.com/gigurra/tofu/cmd/hash"
	"github.com/gigurra/tofu/cmd/head"
	"github.com/gigurra/tofu/cmd/http"
	"github.com/gigurra/tofu/cmd/ip"
	"github.com/gigurra/tofu/cmd/jwt"
	"github.com/gigurra/tofu/cmd/k8s"
	"github.com/gigurra/tofu/cmd/leet"
	"github.com/gigurra/tofu/cmd/lolcat"
	"github.com/gigurra/tofu/cmd/lorem"
	"github.com/gigurra/tofu/cmd/ls"
	"github.com/gigurra/tofu/cmd/magic8"
	"github.com/gigurra/tofu/cmd/mkdir"
	"github.com/gigurra/tofu/cmd/morse"
	"github.com/gigurra/tofu/cmd/mv"
	"github.com/gigurra/tofu/cmd/name"
	"github.com/gigurra/tofu/cmd/nc"
	"github.com/gigurra/tofu/cmd/pick"
	"github.com/gigurra/tofu/cmd/ping"
	"github.com/gigurra/tofu/cmd/pomodoro"
	"github.com/gigurra/tofu/cmd/port"
	"github.com/gigurra/tofu/cmd/proxy"
	"github.com/gigurra/tofu/cmd/ps"
	"github.com/gigurra/tofu/cmd/pwd"
	"github.com/gigurra/tofu/cmd/qr"
	"github.com/gigurra/tofu/cmd/rand"
	"github.com/gigurra/tofu/cmd/reverse"
	"github.com/gigurra/tofu/cmd/rm"
	"github.com/gigurra/tofu/cmd/rmdir"
	"github.com/gigurra/tofu/cmd/screensaver"
	"github.com/gigurra/tofu/cmd/sed2"
	"github.com/gigurra/tofu/cmd/serve"
	"github.com/gigurra/tofu/cmd/sponge"
	"github.com/gigurra/tofu/cmd/standup"
	"github.com/gigurra/tofu/cmd/stopwatch"
	"github.com/gigurra/tofu/cmd/tail"
	"github.com/gigurra/tofu/cmd/tee"
	"github.com/gigurra/tofu/cmd/time"
	"github.com/gigurra/tofu/cmd/touch"
	"github.com/gigurra/tofu/cmd/tree"
	"github.com/gigurra/tofu/cmd/typing"
	"github.com/gigurra/tofu/cmd/uuid"
	"github.com/gigurra/tofu/cmd/uwu"
	"github.com/gigurra/tofu/cmd/watch"
	"github.com/gigurra/tofu/cmd/weather"
	"github.com/gigurra/tofu/cmd/wget"
	"github.com/gigurra/tofu/cmd/which"
	"github.com/spf13/cobra"
)

// Command group IDs
const (
	groupFile         = "file"
	groupNetwork      = "network"
	groupEncoding     = "encoding"
	groupSystem       = "system"
	groupText         = "text"
	groupRandom       = "random"
	groupFun          = "fun"
	groupProductivity = "productivity"
	groupDevOps       = "devops"
)

// withGroup sets the GroupID on a command and returns it
func withGroup(cmd *cobra.Command, group string) *cobra.Command {
	cmd.GroupID = group
	return cmd
}

func main() {
	boa.CmdT[boa.NoParams]{
		Use:     "tofu",
		Short:   "Tools Of Fantastic Utility",
		Version: appVersion(),
		Groups: []*cobra.Group{
			{ID: groupFile, Title: "File Operations:"},
			{ID: groupNetwork, Title: "Network:"},
			{ID: groupEncoding, Title: "Encoding & Crypto:"},
			{ID: groupSystem, Title: "System:"},
			{ID: groupText, Title: "Text Processing:"},
			{ID: groupRandom, Title: "Randomness:"},
			{ID: groupFun, Title: "Fun & Silly:"},
			{ID: groupProductivity, Title: "Productivity:"},
			{ID: groupDevOps, Title: "DevOps:"},
		},
		SubCmds: []*cobra.Command{
			// File Operations
			withGroup(cat.Cmd(), groupFile),
			withGroup(find.Cmd(), groupFile),
			withGroup(grep.Cmd(), groupFile),
			withGroup(sed2.Cmd(), groupFile),
			withGroup(tree.Cmd(), groupFile),
			withGroup(ls.Cmd(), groupFile),
			withGroup(ls.LlCmd(), groupFile),
			withGroup(ls.LaCmd(), groupFile),
			withGroup(mkdir.Cmd(), groupFile),
			withGroup(mv.Cmd(), groupFile),
			withGroup(cp.Cmd(), groupFile),
			withGroup(rm.Cmd(), groupFile),
			withGroup(rmdir.Cmd(), groupFile),
			withGroup(touch.Cmd(), groupFile),
			withGroup(head.Cmd(), groupFile),
			withGroup(tail.Cmd(), groupFile),
			withGroup(diff.Cmd(), groupFile),
			withGroup(du.Cmd(), groupFile),
			withGroup(df.Cmd(), groupFile),

			// Network
			withGroup(serve.Cmd(), groupNetwork),
			withGroup(port.Cmd(), groupNetwork),
			withGroup(ip.Cmd(), groupNetwork),
			withGroup(dns.Cmd(), groupNetwork),
			withGroup(http.Cmd(), groupNetwork),
			withGroup(nc.Cmd(), groupNetwork),
			withGroup(proxy.Cmd(), groupNetwork),
			withGroup(ping.Cmd(), groupNetwork),
			withGroup(wget.Cmd(), groupNetwork),
			withGroup(weather.Cmd(), groupNetwork),

			// Encoding & Crypto
			withGroup(base64.Cmd(), groupEncoding),
			withGroup(crypt.Cmd(), groupEncoding),
			withGroup(jwt.Cmd(), groupEncoding),
			withGroup(hash.Cmd(), groupEncoding),
			withGroup(qr.Cmd(), groupEncoding),
			withGroup(uuid.Cmd(), groupEncoding),
			withGroup(morse.Cmd(), groupEncoding),

			// System
			withGroup(free.Cmd(), groupSystem),
			withGroup(ps.Cmd(), groupSystem),
			withGroup(pwd.Cmd(), groupSystem),
			withGroup(time.Cmd(), groupSystem),
			withGroup(env.Cmd(), groupSystem),
			withGroup(watch.Cmd(), groupSystem),
			withGroup(which.Cmd(), groupSystem),
			withGroup(clip.Cmd(), groupSystem),
			withGroup(cron.Cmd(), groupSystem),
			withGroup(archive.Cmd(), groupSystem),

			// Text Processing
			withGroup(count.Cmd(), groupText),
			withGroup(reverse.Cmd(), groupText),
			withGroup(tee.Cmd(), groupText),
			withGroup(sponge.Cmd(), groupText),
			withGroup(leet.Cmd(), groupText),
			withGroup(lolcat.Cmd(), groupText),
			withGroup(uwu.Cmd(), groupText),
			withGroup(figlet.Cmd(), groupText),

			// Randomness
			withGroup(rand.Cmd(), groupRandom),
			withGroup(coin.Cmd(), groupRandom),
			withGroup(dice.Cmd(), groupRandom),
			withGroup(pick.Cmd(), groupRandom),
			withGroup(name.Cmd(), groupRandom),
			withGroup(lorem.Cmd(), groupRandom),

			// Fun & Silly
			withGroup(bird.Cmd(), groupFun),
			withGroup(blame.Cmd(), groupFun),
			withGroup(excuse.Cmd(), groupFun),
			withGroup(magic8.Cmd(), groupFun),
			withGroup(flip.Cmd(), groupFun),
			withGroup(busy.Cmd(), groupFun),
			withGroup(fortune.Cmd(), groupFun),
			withGroup(cowsay.Cmd(), groupFun),
			withGroup(screensaver.Cmd(), groupFun),
			withGroup(typing.Cmd(), groupFun),

			// Productivity
			withGroup(pomodoro.Cmd(), groupProductivity),
			withGroup(standup.Cmd(), groupProductivity),
			withGroup(calendar.Cmd(), groupProductivity),
			withGroup(clock.Cmd(), groupProductivity),
			withGroup(stopwatch.Cmd(), groupProductivity),

			// DevOps
			withGroup(k8s.Cmd(), groupDevOps),
			withGroup(gh.Cmd(), groupDevOps),
			withGroup(git.Cmd(), groupDevOps),
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
