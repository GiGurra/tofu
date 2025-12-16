package pods

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	FromDeploy    boa.Required[[]string] `short:"d" optional:"true" help:"Filter pods by deployment name (can be repeated)" default:"[]"`
	Labels        []string               `short:"l" optional:"true" help:"Label selector (can be repeated, AND logic)"`
	Names         []string               `short:"n" optional:"true" help:"Pod name pattern filter (substring match, can be repeated, OR logic)"`
	Namespace     boa.Required[string]   `optional:"true" help:"Kubernetes namespace (default: current context)" default:""`
	AllNamespaces bool                   `short:"A" help:"Search pods in all namespaces" default:"false"`
	MaxPods       int                    `help:"Maximum pods to tail simultaneously" default:"10"`
	Tail          int                    `help:"Number of lines to initially read" default:"20"`
	Since         string                 `optional:"true" help:"Only return logs newer than relative duration (e.g., 5m, 1h)"`
	Interval      int                    `help:"Pod discovery poll interval in milliseconds" default:"250"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "pods",
		Short:       "Tail logs from Kubernetes pods",
		Long:        "Continuously tail logs from Kubernetes pods matching the specified criteria. Automatically discovers new pods and handles pod restarts.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {

			params.FromDeploy.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {

				cmdAndArgs := []string{"kubectl", "get", "deployments", "-o", "name"}
				if params.Namespace.Value() != "" {
					cmdAndArgs = append(cmdAndArgs, "-n", params.Namespace.Value())
				}
				if params.AllNamespaces {
					cmdAndArgs = append(cmdAndArgs, "-A")
				}

				res := cmder.New(cmdAndArgs...).
					WithAttemptTimeout(5 * time.Second).
					Run(context.Background())
				if res.Err != nil {
					return nil
				}
				var deployments []string
				lines := strings.Split(strings.TrimSpace(res.StdOut), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					deployName := strings.TrimPrefix(strings.TrimPrefix(line, "deployment/"), "deployment.apps/")
					if strings.HasPrefix(deployName, toComplete) {
						deployments = append(deployments, deployName)
					}
				}
				return deployments
			}

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

func run(params *Params, stdout, stderr io.Writer) error {
	if err := checkKubectl(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(stderr, "\nShutting down...")
		cancel()
	}()

	tailer := &podTailer{
		params:        params,
		stdout:        stdout,
		stderr:        stderr,
		monitoredPods: make(map[string]*podProcess),
		mu:            &sync.Mutex{},
	}

	return tailer.run(ctx)
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

type podProcess struct {
	cancel context.CancelFunc
}

type podTailer struct {
	params        *Params
	stdout        io.Writer
	stderr        io.Writer
	monitoredPods map[string]*podProcess
	mu            *sync.Mutex
}

func (t *podTailer) run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(t.params.Interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.stopAll()
			return nil
		case <-ticker.C:
			if err := t.discoverAndTail(ctx); err != nil {
				fmt.Fprintf(t.stderr, "Discovery error: %v\n", err)
			}
		}
	}
}

func (t *podTailer) stopAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, p := range t.monitoredPods {
		p.cancel()
	}
}

func (t *podTailer) discoverAndTail(ctx context.Context) error {
	pods, err := t.discoverPods(ctx)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Clean up pods that no longer exist
	for podName, proc := range t.monitoredPods {
		found := false
		for _, p := range pods {
			if p == podName {
				found = true
				break
			}
		}
		if !found {
			proc.cancel()
			delete(t.monitoredPods, podName)
		}
	}

	// Start tailing new pods
	for _, podName := range pods {
		if _, exists := t.monitoredPods[podName]; exists {
			continue
		}

		if len(t.monitoredPods) >= t.params.MaxPods {
			fmt.Fprintf(t.stderr, "Max pods (%d) reached, skipping %s\n", t.params.MaxPods, podName)
			continue
		}

		// Check if pod is ready
		if !t.isPodReady(ctx, podName) {
			continue
		}

		podCtx, podCancel := context.WithCancel(ctx)
		t.monitoredPods[podName] = &podProcess{cancel: podCancel}

		fmt.Fprintf(t.stderr, "Tailing %s\n", podName)
		go t.tailPod(podCtx, podName)
	}

	return nil
}

func (t *podTailer) discoverPods(ctx context.Context) ([]string, error) {
	var args []string

	if t.params.AllNamespaces {
		// Use custom columns to get namespace and name together
		args = []string{"get", "pods", "-A", "-o", "custom-columns=NS:.metadata.namespace,NAME:.metadata.name", "--no-headers"}
	} else {
		args = []string{"get", "pods", "-o", "name"}
		if t.params.Namespace.Value() != "" {
			args = append(args, "-n", t.params.Namespace.Value())
		}
	}

	// Add label selectors
	for _, label := range t.params.Labels {
		args = append(args, "-l", label)
	}

	// Add deployment label selectors
	for _, deploy := range t.params.FromDeploy.Value() {
		args = append(args, "-l", fmt.Sprintf("app.kubernetes.io/name=%s", deploy))
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		if result.Combined != "" {
			return nil, fmt.Errorf("kubectl get pods failed: %w\n%s", result.Err, result.Combined)
		}
		return nil, fmt.Errorf("kubectl get pods failed: %w", result.Err)
	}

	var pods []string
	lines := strings.Split(strings.TrimSpace(result.StdOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var podKey string
		if t.params.AllNamespaces {
			// Parse "namespace   podname" format (separated by whitespace)
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			namespace, podName := fields[0], fields[1]
			if t.matchesNameFilter(podName) {
				podKey = namespace + "/" + podName
			}
		} else {
			// kubectl returns "pod/<name>", extract just the name
			podName := strings.TrimPrefix(line, "pod/")
			if t.matchesNameFilter(podName) {
				podKey = podName
			}
		}

		if podKey != "" {
			pods = append(pods, podKey)
		}
	}

	return pods, nil
}

func (t *podTailer) matchesNameFilter(podName string) bool {
	// If no name filters and no deployment filters, match all
	if len(t.params.Names) == 0 && len(t.params.FromDeploy.Value()) == 0 {
		return true
	}

	// Check name patterns (OR logic)
	for _, pattern := range t.params.Names {
		if strings.Contains(podName, pattern) {
			return true
		}
	}

	// If we have deployment filters but no name filters, the label selector handles it
	if len(t.params.Names) == 0 && len(t.params.FromDeploy.Value()) > 0 {
		return true
	}

	// Also check if pod name starts with deployment name (fallback matching)
	for _, deploy := range t.params.FromDeploy.Value() {
		if strings.HasPrefix(podName, deploy+"-") {
			return true
		}
	}

	return len(t.params.Names) == 0
}

// parsePodKey extracts namespace and pod name from a pod key.
// For all-namespaces mode, key format is "namespace/podname".
// Otherwise, key is just "podname".
func (t *podTailer) parsePodKey(podKey string) (namespace, podName string) {
	if t.params.AllNamespaces {
		parts := strings.SplitN(podKey, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return t.params.Namespace.Value(), podKey
}

func (t *podTailer) isPodReady(ctx context.Context, podKey string) bool {
	namespace, podName := t.parsePodKey(podKey)
	args := []string{"logs", "--tail=1", podName}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)
	return result.Err == nil
}

func (t *podTailer) tailPod(ctx context.Context, podKey string) {
	namespace, podName := t.parsePodKey(podKey)
	args := []string{"logs", "-f", "--tail", fmt.Sprintf("%d", t.params.Tail), podName}

	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	if t.params.Since != "" {
		args = append(args, "--since", t.params.Since)
	}

	// Create line-prefixing writers for stdout and stderr
	// Include namespace in prefix when using all-namespaces
	var prefix string
	if t.params.AllNamespaces {
		prefix = fmt.Sprintf("[%s/%s] ", namespace, podName)
	} else {
		prefix = fmt.Sprintf("[%s] ", podName)
	}
	stdoutWriter := &prefixWriter{
		prefix: prefix,
		dest:   t.stdout,
	}
	stderrWriter := &prefixWriter{
		prefix: prefix,
		dest:   t.stdout, // Send stderr to stdout as well, like ktail does
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithStdOut(stdoutWriter).
		WithStdErr(stderrWriter).
		Run(ctx)

	t.mu.Lock()
	if result.Err != nil && ctx.Err() == nil {
		// Pod terminated with error (not due to context cancellation), allow re-tailing
		delete(t.monitoredPods, podKey)
		if result.Combined != "" {
			fmt.Fprintf(t.stderr, "Pod %s terminated: %v\n%s", podKey, result.Err, result.Combined)
		} else {
			fmt.Fprintf(t.stderr, "Pod %s terminated: %v\n", podKey, result.Err)
		}
	}
	t.mu.Unlock()
}

// prefixWriter wraps an io.Writer and prefixes each line with a given string
type prefixWriter struct {
	prefix string
	dest   io.Writer
	buf    bytes.Buffer
	mu     sync.Mutex
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n = len(p)
	w.buf.Write(p)

	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			// No complete line yet, put back what we read
			w.buf.WriteString(line)
			break
		}
		// Write the prefixed line
		fmt.Fprint(w.dest, w.prefix+line)
	}

	return n, nil
}

// Flush writes any remaining buffered data
func (w *prefixWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() > 0 {
		remaining := w.buf.String()
		w.buf.Reset()
		if remaining != "" {
			fmt.Fprintln(w.dest, w.prefix+remaining)
		}
	}
}

// Close implements io.Closer for prefixWriter
func (w *prefixWriter) Close() error {
	// Flush any remaining content in the buffer
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() > 0 {
		remaining := w.buf.String()
		w.buf.Reset()
		// Use scanner to handle any partial lines
		scanner := bufio.NewScanner(strings.NewReader(remaining))
		for scanner.Scan() {
			fmt.Fprintln(w.dest, w.prefix+scanner.Text())
		}
	}
	return nil
}
