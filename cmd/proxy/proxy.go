package proxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Listen string `pos:"true" help:"Address to listen on (e.g. 0.0.0.0:8443)"`
	Target string `pos:"true" help:"Address to forward to (e.g. localhost:8443)"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "proxy <listen-addr> <target-addr>",
		Short: "TCP proxy/port forwarder",
		Long: `Forward TCP connections from a listen address to a target address.

Useful for exposing WSL services on Windows LAN interfaces, or any TCP forwarding.

Example:
  tofu proxy 0.0.0.0:8443 localhost:8443`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := run(params); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func run(params *Params) error {
	ln, err := net.Listen("tcp", params.Listen)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", params.Listen, err)
	}
	defer ln.Close()

	fmt.Printf("Proxying %s -> %s\n", params.Listen, params.Target)

	var connCount atomic.Int64

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		n := connCount.Add(1)
		fmt.Printf("[%d] %s connected\n", n, conn.RemoteAddr())

		go func(id int64) {
			defer conn.Close()

			target, err := net.Dial("tcp", params.Target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%d] connect to %s: %v\n", id, params.Target, err)
				return
			}
			defer target.Close()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				io.Copy(target, conn)
				target.(*net.TCPConn).CloseWrite()
			}()

			go func() {
				defer wg.Done()
				io.Copy(conn, target)
				conn.(*net.TCPConn).CloseWrite()
			}()

			wg.Wait()
			fmt.Printf("[%d] disconnected\n", id)
		}(n)
	}
}
