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
	testDaemonSet = "nginx-ds-test"
)

func setupDaemonSet(ns string) error {
	ctx := context.Background()

	// Create DaemonSet
	dsYaml := fmt.Sprintf(`
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: %s
  namespace: %s
spec:
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
`, testDaemonSet, ns, testDaemonSet, testDaemonSet)

	result := cmder.New("kubectl", "apply", "-f", "-").
		WithStdIn(strings.NewReader(dsYaml)).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("failed to create daemonset: %w\n%s", result.Err, result.Combined)
	}

	// Wait for pod to be ready
	result = cmder.New("kubectl", "wait", "--for=condition=ready", "pod",
		"-l", fmt.Sprintf("app=%s", testDaemonSet),
		"-n", ns,
		"--timeout=60s").
		WithAttemptTimeout(70 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("daemonset pod not ready: %w\n%s", result.Err, result.Combined)
	}

	return nil
}

func teardownDaemonSet(ns string) {
	ctx := context.Background()
	_ = cmder.New("kubectl", "delete", "daemonset", testDaemonSet, "-n", ns, "--ignore-not-found").
		WithAttemptTimeout(30 * time.Second).
		Run(ctx)
}

func TestPortForward_DaemonSet(t *testing.T) {
	// Setup DaemonSet in test namespace
	if err := setupDaemonSet(testNamespace1); err != nil {
		t.Fatalf("Failed to setup DaemonSet: %v", err)
	}
	defer teardownDaemonSet(testNamespace1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source:    Source{Kind: SourceDaemonSet, Name: testDaemonSet},
		Namespace: testNamespace1,
		Ports:     []string{"18091:80"},
		Keepalive: false,
	}

	// Start port-forward in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- runWithConfig(ctx, cfg, io.Discard, os.Stderr)
	}()

	// Wait for port-forward to establish
	time.Sleep(2 * time.Second)

	// Try to connect to nginx
	resp, err := http.Get("http://localhost:18091")
	if err != nil {
		t.Fatalf("Failed to connect to nginx via DaemonSet port-forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Successfully connected to nginx via DaemonSet port-forward")
	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestGetDaemonSetSelectorLabels(t *testing.T) {
	// Setup DaemonSet
	if err := setupDaemonSet(testNamespace1); err != nil {
		t.Fatalf("Failed to setup DaemonSet: %v", err)
	}
	defer teardownDaemonSet(testNamespace1)

	ctx := context.Background()

	labels, err := getDaemonSetSelectorLabels(ctx, testDaemonSet, testNamespace1)
	if err != nil {
		t.Fatalf("Failed to get selector labels: %v", err)
	}

	if labels == nil {
		t.Fatal("Expected to get selector labels, got nil")
	}

	if _, ok := labels["app"]; !ok {
		t.Errorf("Expected 'app' label in selector, got: %v", labels)
	}

	if labels["app"] != testDaemonSet {
		t.Errorf("Expected app=%s, got app=%s", testDaemonSet, labels["app"])
	}

	t.Logf("Got DaemonSet selector labels: %v", labels)
}
