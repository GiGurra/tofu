package proxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Listen        string `pos:"true" help:"Address to listen on (e.g. 0.0.0.0:8443)"`
	Target        string `pos:"true" help:"Address to forward to (e.g. localhost:8443)"`
	ConnectTimeout int64 `short:"t" help:"Connect timeout in ms (0=no timeout)" default:"5000"`
	IdleTimeout    int64 `short:"i" help:"Idle timeout in ms, close if no data (0=no timeout)" default:"0"`
	Retries        int   `short:"r" help:"Connection retries (-1=infinite, 0=no retry)" default:"0"`
	RetryInterval  int64 `help:"Retry interval in ms" default:"1000"`
	MaxConns       int   `short:"m" help:"Max concurrent connections (0=unlimited)" default:"0"`
	Verbose        bool  `short:"v" help:"Verbose logging" default:"false"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "proxy",
		Short: "TCP proxy/port forwarder",
		Long: `Forward TCP connections from a listen address to a target address.

Useful for exposing WSL services on Windows LAN interfaces, or any TCP forwarding.

Example:
  tofu proxy 0.0.0.0:8443 localhost:8443
  tofu proxy -t 10000 -i 60000 -r 3 0.0.0.0:8443 localhost:8443`,
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
	if params.Verbose {
		fmt.Printf("  connect-timeout: %dms, idle-timeout: %dms, retries: %d, retry-interval: %dms, max-conns: %d\n",
			params.ConnectTimeout, params.IdleTimeout, params.Retries, params.RetryInterval, params.MaxConns)
	}

	var connCount atomic.Int64
	var activeConns atomic.Int64

	var sem chan struct{}
	if params.MaxConns > 0 {
		sem = make(chan struct{}, params.MaxConns)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}

		id := connCount.Add(1)

		if sem != nil {
			select {
			case sem <- struct{}{}:
			default:
				fmt.Fprintf(os.Stderr, "[%d] rejected %s (max connections: %d)\n", id, conn.RemoteAddr(), params.MaxConns)
				conn.Close()
				continue
			}
		}

		active := activeConns.Add(1)
		fmt.Printf("[%d] %s connected (active: %d)\n", id, conn.RemoteAddr(), active)

		go func() {
			start := time.Now()
			defer func() {
				conn.Close()
				remaining := activeConns.Add(-1)
				if params.Verbose {
					fmt.Printf("[%d] disconnected after %s (active: %d)\n", id, time.Since(start).Round(time.Millisecond), remaining)
				} else {
					fmt.Printf("[%d] disconnected (active: %d)\n", id, remaining)
				}
				if sem != nil {
					<-sem
				}
			}()

			target, err := dialWithRetry(params, id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[%d] connect to %s: %v\n", id, params.Target, err)
				return
			}
			defer target.Close()

			idleTimeout := time.Duration(params.IdleTimeout) * time.Millisecond

			var wg sync.WaitGroup
			wg.Add(2)

			var bytesSent, bytesRecv atomic.Int64

			// client -> target
			go func() {
				defer wg.Done()
				n := copyWithIdleTimeout(target, conn, idleTimeout)
				bytesSent.Store(n)
				if tc, ok := target.(*net.TCPConn); ok {
					tc.CloseWrite()
				}
			}()

			// target -> client
			go func() {
				defer wg.Done()
				n := copyWithIdleTimeout(conn, target, idleTimeout)
				bytesRecv.Store(n)
				if tc, ok := conn.(*net.TCPConn); ok {
					tc.CloseWrite()
				}
			}()

			wg.Wait()

			if params.Verbose {
				fmt.Printf("[%d] sent %s, received %s\n", id, formatBytes(bytesSent.Load()), formatBytes(bytesRecv.Load()))
			}
		}()
	}
}

func dialWithRetry(params *Params, id int64) (net.Conn, error) {
	var dialer net.Dialer
	if params.ConnectTimeout > 0 {
		dialer.Timeout = time.Duration(params.ConnectTimeout) * time.Millisecond
	}

	maxAttempts := params.Retries + 1
	if params.Retries < 0 {
		maxAttempts = -1
	}

	for attempt := 1; maxAttempts < 0 || attempt <= maxAttempts; attempt++ {
		conn, err := dialer.Dial("tcp", params.Target)
		if err == nil {
			if params.Verbose && attempt > 1 {
				fmt.Printf("[%d] connected to target on attempt %d\n", id, attempt)
			}
			return conn, nil
		}

		if maxAttempts > 0 && attempt >= maxAttempts {
			return nil, err
		}

		if params.Verbose {
			fmt.Fprintf(os.Stderr, "[%d] attempt %d failed: %v, retrying in %dms\n",
				id, attempt, err, params.RetryInterval)
		}
		time.Sleep(time.Duration(params.RetryInterval) * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to %s", params.Target)
}

func copyWithIdleTimeout(dst, src net.Conn, idleTimeout time.Duration) int64 {
	if idleTimeout <= 0 {
		n, _ := io.Copy(dst, src)
		return n
	}

	buf := make([]byte, 32*1024)
	var total int64
	for {
		src.SetReadDeadline(time.Now().Add(idleTimeout))
		nr, readErr := src.Read(buf)
		if nr > 0 {
			nw, writeErr := dst.Write(buf[:nr])
			total += int64(nw)
			if writeErr != nil {
				break
			}
		}
		if readErr != nil {
			break
		}
	}
	return total
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
