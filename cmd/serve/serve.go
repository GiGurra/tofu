package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Dir     string `pos:"true" optional:"true" help:"Directory to serve." default:"."`
	Port    int    `short:"p" help:"Port to listen on." default:"8080"`
	Host    string `help:"Host interface to bind to." default:"localhost"`
	SpaMode bool   `help:"Enable Single Page Application mode (redirect 404 to index.html)." default:"false"`
	NoCache bool   `help:"Disable browser caching." default:"false"`

	ReadTimeoutMillis  int64 `help:"Maximum duration for reading the entire request, including the body (ms)." default:"5000"`
	WriteTimeoutMillis int64 `help:"Maximum duration before timing out writes of the response (ms)." default:"10000"`
	IdleTimeoutMillis  int64 `help:"Maximum amount of time to wait for the next request when keep-alives are enabled (ms)." default:"120000"`
	MaxHeaderBytes     int   `help:"Maximum number of bytes the server will read parsing the request header's keys and values." default:"1048576"` // 1MB
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "serve",
		Short:       "Instant static file server",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(cmd.Context(), params); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "serve: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(ctx context.Context, params *Params) error {
	absDir, err := filepath.Abs(params.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory %s: %w", params.Dir, err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	fs := http.FileServer(http.Dir(absDir))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Headers
		if params.NoCache {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}

		// SPA handling
		if params.SpaMode {
			// Check if file exists, if not serve index.html
			fPath := filepath.Join(absDir, r.URL.Path)
			// Basic check: if it has no extension and doesn't exist, or if it explicitly doesn't exist and isn't an asset
			// A simple robust way for SPA:
			// If file exists, serve it.
			// If not, serve index.html.
			if _, err := os.Stat(fPath); os.IsNotExist(err) {
				r.URL.Path = "/"
			}
		}

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		fs.ServeHTTP(rw, r)

		// Log
		duration := time.Since(start)
		fmt.Printf("[%d] %s %s (%v)\n", rw.status, r.Method, r.URL.Path, duration)
	})

	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	server := &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    time.Duration(params.ReadTimeoutMillis) * time.Millisecond,
		WriteTimeout:   time.Duration(params.WriteTimeoutMillis) * time.Millisecond,
		IdleTimeout:    time.Duration(params.IdleTimeoutMillis) * time.Millisecond,
		MaxHeaderBytes: params.MaxHeaderBytes,
	}

	// Handle graceful shutdown
	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("Serving %s at http://%s\n", absDir, addr)
		if params.SpaMode {
			fmt.Println("SPA Mode enabled (redirecting 404s to index.html)")
		}
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		return nil
	case err := <-serverErr:
		return err
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
