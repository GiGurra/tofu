package portforward

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type Params struct {
	FromDeploy boa.Required[string] `short:"d" optional:"true" help:"Deployment name to port-forward to"`
	FromSts    boa.Required[string] `optional:"true" help:"StatefulSet name to port-forward to"`
	FromDs     boa.Required[string] `optional:"true" help:"DaemonSet name to port-forward to"`
	FromSvc    boa.Required[string] `optional:"true" help:"Service name to port-forward to"`
	Ports      []string             `pos:"true" help:"Ports to forward (format: [local:]remote, e.g., 8080:80 or just 80)"`
	Namespace  boa.Required[string] `short:"n" optional:"true" help:"Kubernetes namespace (default: current context)" default:""`
	Keepalive  bool                 `short:"k" help:"Keep trying to reconnect when connection is lost" default:"true"`
}

// Config holds the resolved configuration for port-forwarding (used internally and for testing)
type Config struct {
	Source           Source
	Ports            []string
	Namespace        string
	Keepalive        bool
	PodCheckInterval time.Duration // How often to check if the pod is still running (0 = use default of 2s)
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "port-forward",
		Short:       "Port-forward to a pod from a workload with auto-reconnect",
		Long:        "Port-forward to a running pod from a deployment, statefulset, daemonset, or service. Automatically reconnects when the connection is lost or the pod terminates.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			// Deployment completions
			params.FromDeploy.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				return lo.Filter(findDeployments(context.Background(), params.Namespace.Value()), func(item string, _ int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			// StatefulSet completions
			params.FromSts.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				return lo.Filter(findStatefulSets(context.Background(), params.Namespace.Value()), func(item string, _ int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			// DaemonSet completions
			params.FromDs.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				return lo.Filter(findDaemonSets(context.Background(), params.Namespace.Value()), func(item string, _ int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			// Service completions
			params.FromSvc.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				return lo.Filter(findServices(context.Background(), params.Namespace.Value()), func(item string, _ int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			// Namespace completions
			params.Namespace.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				res := cmder.New("kubectl", "get", "namespaces", "-o", "name").
					WithAttemptTimeout(5 * time.Second).
					Run(context.Background())
				if res.Err != nil {
					return nil
				}
				var namespaces []string
				lines := strings.Split(strings.TrimSpace(res.StdOut), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					nsName := strings.TrimPrefix(line, "namespace/")
					if strings.HasPrefix(nsName, toComplete) {
						namespaces = append(namespaces, nsName)
					}
				}
				return namespaces
			}

			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := run(params, os.Stdout, os.Stderr); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func (p *Params) toConfig() (Config, error) {
	source, err := p.getSource()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Source:    source,
		Ports:     p.Ports,
		Namespace: p.Namespace.Value(),
		Keepalive: p.Keepalive,
	}, nil
}

func (p *Params) getSource() (Source, error) {
	var sources []Source

	if p.FromDeploy.Value() != "" {
		sources = append(sources, Source{Kind: SourceDeployment, Name: p.FromDeploy.Value()})
	}
	if p.FromSts.Value() != "" {
		sources = append(sources, Source{Kind: SourceStatefulSet, Name: p.FromSts.Value()})
	}
	if p.FromDs.Value() != "" {
		sources = append(sources, Source{Kind: SourceDaemonSet, Name: p.FromDs.Value()})
	}
	if p.FromSvc.Value() != "" {
		sources = append(sources, Source{Kind: SourceService, Name: p.FromSvc.Value()})
	}

	if len(sources) == 0 {
		return Source{}, fmt.Errorf("one of --from-deploy, --from-sts, --from-ds, or --from-svc is required")
	}
	if len(sources) > 1 {
		return Source{}, fmt.Errorf("only one of --from-deploy, --from-sts, --from-ds, or --from-svc can be specified")
	}

	return sources[0], nil
}

func run(params *Params, stdout, stderr io.Writer) error {
	cfg, err := params.toConfig()
	if err != nil {
		return err
	}
	return runWithConfig(context.Background(), cfg, stdout, stderr)
}

func runWithConfig(parentCtx context.Context, cfg Config, stdout, stderr io.Writer) error {
	if err := checkKubectl(); err != nil {
		return err
	}

	if cfg.Source.IsEmpty() {
		return fmt.Errorf("no source specified")
	}

	if len(cfg.Ports) == 0 {
		return fmt.Errorf("at least one port is required")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			fmt.Fprintln(stderr, "\nShutting down...")
			cancel()
		case <-ctx.Done():
			// Parent context cancelled, clean up signal handler
		}
	}()

	forwarder := &portForwarder{
		config: cfg,
		stdout: stdout,
		stderr: stderr,
	}

	return forwarder.run(ctx)
}

func checkKubectl() error {
	result := cmder.New("kubectl", "version", "--client").
		WithAttemptTimeout(5 * time.Second).
		Run(context.Background())
	if result.Err != nil {
		if result.Combined != "" {
			return fmt.Errorf("kubectl not found or not working: %w\n%s", result.Err, result.Combined)
		}
		return fmt.Errorf("kubectl not found or not working: %w", result.Err)
	}
	return nil
}

type portForwarder struct {
	config Config
	stdout io.Writer
	stderr io.Writer
}

func (f *portForwarder) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Find a running pod
		pod, err := f.findRunningPod(ctx)
		if err != nil {
			fmt.Fprintf(f.stderr, "Error finding pod: %v\n", err)
			if !f.config.Keepalive {
				return err
			}
			// Wait before retrying
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(1 * time.Second):
				continue
			}
		}

		if pod == "" {
			fmt.Fprintf(f.stderr, "No running pods found for %s, waiting...\n", f.config.Source)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(1 * time.Second):
				continue
			}
		}

		// Port forward to the pod
		startTime := time.Now()
		fmt.Fprintf(f.stderr, "Port-forwarding to pod %s (ports: %s)\n", pod, strings.Join(f.config.Ports, ", "))
		err = f.portForward(ctx, pod)
		duration := time.Since(startTime)

		if ctx.Err() != nil {
			// Context was cancelled, exit cleanly
			return nil
		}

		if err != nil {
			fmt.Fprintf(f.stderr, "Port-forward to %s terminated: %v\n", pod, err)
		} else {
			fmt.Fprintf(f.stderr, "Port-forward to %s terminated\n", pod)
		}

		if !f.config.Keepalive {
			return err
		}

		// If connection was lost quicker than 5 seconds, wait 1 second before reconnecting
		if duration < 5*time.Second {
			fmt.Fprintf(f.stderr, "Connection lost quickly, waiting 1s before reconnecting...\n")
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(1 * time.Second):
			}
		}
	}
}

func (f *portForwarder) findRunningPod(ctx context.Context) (string, error) {
	// Get selector labels from the source
	selectorLabels, err := f.config.Source.GetSelectorLabels(ctx, f.config.Namespace)
	if err != nil {
		return "", err
	}

	if len(selectorLabels) == 0 {
		return "", fmt.Errorf("no selector labels found for %s", f.config.Source)
	}

	// Build label selector
	var labelParts []string
	for key, value := range selectorLabels {
		labelParts = append(labelParts, fmt.Sprintf("%s=%s", key, value))
	}
	labelSelector := strings.Join(labelParts, ",")

	// Get pods with Running status
	args := []string{"get", "pods", "-l", labelSelector, "-o", "jsonpath={.items[*].metadata.name}", "--field-selector=status.phase=Running"}
	if f.config.Namespace != "" {
		args = append(args, "-n", f.config.Namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return "", fmt.Errorf("failed to get pods: %w", result.Err)
	}

	pods := strings.Fields(strings.TrimSpace(result.StdOut))
	if len(pods) == 0 {
		return "", nil
	}

	// Return the first running pod
	return pods[0], nil
}

func (f *portForwarder) portForward(ctx context.Context, pod string) error {
	// Create a cancellable context for this port-forward session
	pfCtx, pfCancel := context.WithCancel(ctx)
	defer pfCancel()

	// Start pod monitor in background if keepalive is enabled
	if f.config.Keepalive {
		go f.monitorPod(pfCtx, pfCancel, pod)
	}

	args := []string{"port-forward", pod}
	args = append(args, f.config.Ports...)
	if f.config.Namespace != "" {
		args = append(args, "-n", f.config.Namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithStdOut(f.stdout).
		WithStdErr(f.stderr).
		Run(pfCtx)

	return result.Err
}

// monitorPod periodically checks if the pod is still running and cancels the context if not
func (f *portForwarder) monitorPod(ctx context.Context, cancel context.CancelFunc, pod string) {
	interval := f.config.PodCheckInterval
	if interval == 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !f.isPodRunning(ctx, pod) {
				fmt.Fprintf(f.stderr, "Pod %s is no longer running, triggering reconnect...\n", pod)
				cancel()
				return
			}
		}
	}
}

// isPodRunning checks if a specific pod is still in Running phase
func (f *portForwarder) isPodRunning(ctx context.Context, pod string) bool {
	args := []string{"get", "pod", pod, "-o", "jsonpath={.status.phase}"}
	if f.config.Namespace != "" {
		args = append(args, "-n", f.config.Namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)

	if result.Err != nil {
		// Pod might not exist anymore
		return false
	}

	phase := strings.TrimSpace(result.StdOut)
	return phase == "Running"
}
