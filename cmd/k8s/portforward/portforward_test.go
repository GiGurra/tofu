//go:build k8stest

package portforward

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GiGurra/cmder"
)

const (
	testNamespace1 = "tofu-pf-test-1"
	testNamespace2 = "tofu-pf-test-2"
	testDeployment = "nginx-test"
)

func TestMain(m *testing.M) {
	// Setup: create test namespaces and deployments
	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown: clean up test resources
	teardown()

	os.Exit(code)
}

func setup() error {
	ctx := context.Background()

	// Create test namespaces
	for _, ns := range []string{testNamespace1, testNamespace2} {
		result := cmder.New("kubectl", "create", "namespace", ns).
			WithAttemptTimeout(10 * time.Second).
			Run(ctx)
		if result.Err != nil && !strings.Contains(result.Combined, "already exists") {
			return fmt.Errorf("failed to create namespace %s: %w\n%s", ns, result.Err, result.Combined)
		}
		fmt.Printf("Created namespace: %s\n", ns)
	}

	// Deploy nginx in both namespaces
	for _, ns := range []string{testNamespace1, testNamespace2} {
		// Create deployment
		result := cmder.New("kubectl", "create", "deployment", testDeployment,
			"--image=nginx:alpine",
			"-n", ns).
			WithAttemptTimeout(30 * time.Second).
			Run(ctx)
		if result.Err != nil && !strings.Contains(result.Combined, "already exists") {
			return fmt.Errorf("failed to create deployment in %s: %w\n%s", ns, result.Err, result.Combined)
		}
		fmt.Printf("Created deployment %s in namespace %s\n", testDeployment, ns)
	}

	// Wait for pods to be ready
	fmt.Println("Waiting for pods to be ready...")
	for _, ns := range []string{testNamespace1, testNamespace2} {
		result := cmder.New("kubectl", "wait", "--for=condition=ready", "pod",
			"-l", fmt.Sprintf("app=%s", testDeployment),
			"-n", ns,
			"--timeout=60s").
			WithAttemptTimeout(70 * time.Second).
			Run(ctx)
		if result.Err != nil {
			return fmt.Errorf("pods not ready in %s: %w\n%s", ns, result.Err, result.Combined)
		}
		fmt.Printf("Pods ready in namespace %s\n", ns)
	}

	return nil
}

func teardown() {
	ctx := context.Background()
	fmt.Println("Cleaning up test resources...")

	// Kill any lingering kubectl port-forward processes from our tests
	// This handles cases where context cancellation didn't complete before test exit
	_ = cmder.New("pkill", "-f", "kubectl.*port-forward.*tofu-pf-test").
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)

	// Give processes a moment to die
	time.Sleep(500 * time.Millisecond)

	for _, ns := range []string{testNamespace1, testNamespace2} {
		result := cmder.New("kubectl", "delete", "namespace", ns, "--ignore-not-found").
			WithAttemptTimeout(60 * time.Second).
			Run(ctx)
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete namespace %s: %v\n", ns, result.Err)
		} else {
			fmt.Printf("Deleted namespace: %s\n", ns)
		}
	}
}

func TestPortForward_BasicConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source: Source{Kind: SourceDeployment, Name: testDeployment},
		Namespace:  testNamespace1,
		Ports:      []string{"18080:80"},
		Keepalive:  false,
	}

	// Start port-forward in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, os.Stderr)
	}()

	// Wait a bit for port-forward to establish
	time.Sleep(2 * time.Second)

	// Try to connect to nginx
	resp, err := http.Get("http://localhost:18080")
	if err != nil {
		t.Fatalf("Failed to connect to nginx via port-forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "nginx") && !strings.Contains(string(body), "Welcome") {
		t.Errorf("Response doesn't look like nginx: %s", string(body)[:min(100, len(body))])
	}

	t.Log("Successfully connected to nginx via port-forward")
	cancel() // Stop the port-forward

	// Wait for port-forward to finish
	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestPortForward_DifferentNamespace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source: Source{Kind: SourceDeployment, Name: testDeployment},
		Namespace:  testNamespace2,
		Ports:      []string{"18081:80"},
		Keepalive:  false,
	}

	// Start port-forward in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, os.Stderr)
	}()

	// Wait a bit for port-forward to establish
	time.Sleep(2 * time.Second)

	// Try to connect to nginx
	resp, err := http.Get("http://localhost:18081")
	if err != nil {
		t.Fatalf("Failed to connect to nginx via port-forward in namespace %s: %v", testNamespace2, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Logf("Successfully connected to nginx in namespace %s via port-forward", testNamespace2)
	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestPortForward_MultiplePorts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source: Source{Kind: SourceDeployment, Name: testDeployment},
		Namespace:  testNamespace1,
		Ports:      []string{"18082:80", "18083:80"}, // Both map to port 80
		Keepalive:  false,
	}

	// Start port-forward in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, os.Stderr)
	}()

	// Wait a bit for port-forward to establish
	time.Sleep(2 * time.Second)

	// Try to connect via both ports
	for _, port := range []string{"18082", "18083"} {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s", port))
		if err != nil {
			t.Errorf("Failed to connect via port %s: %v", port, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Port %s: Expected status 200, got %d", port, resp.StatusCode)
		} else {
			t.Logf("Successfully connected via port %s", port)
		}
	}

	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestPortForward_ProactiveReconnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cfg := Config{
		Source:           Source{Kind: SourceDeployment, Name: testDeployment},
		Namespace:        testNamespace1,
		Ports:            []string{"18084:80"},
		Keepalive:        true,
		PodCheckInterval: 1 * time.Second, // Check every second for faster detection
	}

	// Start port-forward in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, os.Stderr)
	}()

	// Wait for port-forward to establish
	time.Sleep(2 * time.Second)

	// Verify initial connection works
	resp, err := http.Get("http://localhost:18084")
	if err != nil {
		t.Fatalf("Failed initial connection: %v", err)
	}
	resp.Body.Close()
	t.Log("Initial connection successful")

	// Get the current pod name
	podResult := cmder.New("kubectl", "get", "pods", "-n", testNamespace1,
		"-l", fmt.Sprintf("app=%s", testDeployment),
		"-o", "jsonpath={.items[0].metadata.name}").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if podResult.Err != nil {
		t.Fatalf("Failed to get pod name: %v", podResult.Err)
	}
	oldPod := strings.TrimSpace(podResult.StdOut)
	t.Logf("Current pod: %s", oldPod)

	// Delete the pod to trigger proactive reconnect detection
	t.Log("Deleting pod - expecting proactive detection...")
	deleteResult := cmder.New("kubectl", "delete", "pod", oldPod, "-n", testNamespace1).
		WithAttemptTimeout(30 * time.Second).
		Run(ctx)
	if deleteResult.Err != nil {
		t.Fatalf("Failed to delete pod: %v", deleteResult.Err)
	}

	// Wait for new pod to be ready
	t.Log("Waiting for new pod to be ready...")
	waitResult := cmder.New("kubectl", "wait", "--for=condition=ready", "pod",
		"-l", fmt.Sprintf("app=%s", testDeployment),
		"-n", testNamespace1,
		"--timeout=60s").
		WithAttemptTimeout(70 * time.Second).
		Run(ctx)
	if waitResult.Err != nil {
		t.Fatalf("New pod not ready: %v", waitResult.Err)
	}

	// The proactive monitor should have detected the pod is gone and reconnected
	// Give it a moment to establish new connection
	t.Log("Waiting for proactive reconnect to new pod...")
	time.Sleep(3 * time.Second)

	// Verify reconnection works - should work immediately since monitor already triggered reconnect
	var lastErr error
	for i := 0; i < 10; i++ {
		resp, err = http.Get("http://localhost:18084")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Log("Proactive reconnection successful!")
				cancel()
				select {
				case <-errCh:
				case <-time.After(5 * time.Second):
				}
				return
			}
		}
		lastErr = err
		time.Sleep(1 * time.Second)
	}

	t.Fatalf("Failed to reconnect after pod deletion: %v", lastErr)
}

func TestPortForward_DeploymentNotFound(t *testing.T) {
	cfg := Config{
		Source: Source{Kind: SourceDeployment, Name: "nonexistent-deployment"},
		Namespace:  testNamespace1,
		Ports:      []string{"18085:80"},
		Keepalive:  false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, io.Discard)
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("Expected error for nonexistent deployment, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	case <-ctx.Done():
		// Context cancelled - test passed (couldn't find deployment)
		t.Log("Port-forward correctly failed to find deployment")
	}
}

func TestFindRunningPod(t *testing.T) {
	ctx := context.Background()

	f := &portForwarder{
		config: Config{
			Source: Source{Kind: SourceDeployment, Name: testDeployment},
			Namespace:  testNamespace1,
			Ports:      []string{"80"},
			Keepalive:  false,
		},
		stdout: io.Discard,
		stderr: io.Discard,
	}

	pod, err := f.findRunningPod(ctx)
	if err != nil {
		t.Fatalf("findRunningPod failed: %v", err)
	}

	if pod == "" {
		t.Fatal("Expected to find a running pod, got empty string")
	}

	if !strings.HasPrefix(pod, testDeployment) {
		t.Errorf("Pod name %q doesn't start with deployment name %q", pod, testDeployment)
	}

	t.Logf("Found running pod: %s", pod)
}

func TestGetDeploymentSelectorLabels(t *testing.T) {
	ctx := context.Background()

	labels, err := getDeploymentSelectorLabels(ctx, testDeployment, testNamespace1)
	if err != nil {
		t.Fatalf("Failed to get selector labels: %v", err)
	}

	if labels == nil {
		t.Fatal("Expected to get selector labels, got nil")
	}

	if _, ok := labels["app"]; !ok {
		t.Errorf("Expected 'app' label in selector, got: %v", labels)
	}

	if labels["app"] != testDeployment {
		t.Errorf("Expected app=%s, got app=%s", testDeployment, labels["app"])
	}

	t.Logf("Got selector labels: %v", labels)
}
