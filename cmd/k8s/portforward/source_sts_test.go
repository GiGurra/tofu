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
	testStatefulSet = "nginx-sts-test"
)

func setupStatefulSet(ns string) error {
	ctx := context.Background()

	// Create a headless service (required for StatefulSet)
	svcYaml := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  clusterIP: None
  selector:
    app: %s
  ports:
  - port: 80
    name: web
`, testStatefulSet, ns, testStatefulSet)

	result := cmder.New("kubectl", "apply", "-f", "-").
		WithStdIn(strings.NewReader(svcYaml)).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("failed to create headless service: %w\n%s", result.Err, result.Combined)
	}

	// Create StatefulSet
	stsYaml := fmt.Sprintf(`
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: %s
  namespace: %s
spec:
  serviceName: %s
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
`, testStatefulSet, ns, testStatefulSet, testStatefulSet, testStatefulSet)

	result = cmder.New("kubectl", "apply", "-f", "-").
		WithStdIn(strings.NewReader(stsYaml)).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("failed to create statefulset: %w\n%s", result.Err, result.Combined)
	}

	// Wait for pod to be ready
	result = cmder.New("kubectl", "wait", "--for=condition=ready", "pod",
		"-l", fmt.Sprintf("app=%s", testStatefulSet),
		"-n", ns,
		"--timeout=60s").
		WithAttemptTimeout(70 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("statefulset pod not ready: %w\n%s", result.Err, result.Combined)
	}

	return nil
}

func teardownStatefulSet(ns string) {
	ctx := context.Background()
	_ = cmder.New("kubectl", "delete", "statefulset", testStatefulSet, "-n", ns, "--ignore-not-found").
		WithAttemptTimeout(30 * time.Second).
		Run(ctx)
	_ = cmder.New("kubectl", "delete", "service", testStatefulSet, "-n", ns, "--ignore-not-found").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
}

func TestPortForward_StatefulSet(t *testing.T) {
	// Setup StatefulSet in test namespace
	if err := setupStatefulSet(testNamespace1); err != nil {
		t.Fatalf("Failed to setup StatefulSet: %v", err)
	}
	defer teardownStatefulSet(testNamespace1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source:    Source{Kind: SourceStatefulSet, Name: testStatefulSet},
		Namespace: testNamespace1,
		Ports:     []string{"18090:80"},
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
	resp, err := http.Get("http://localhost:18090")
	if err != nil {
		t.Fatalf("Failed to connect to nginx via StatefulSet port-forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Successfully connected to nginx via StatefulSet port-forward")
	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestGetStatefulSetSelectorLabels(t *testing.T) {
	// Setup StatefulSet
	if err := setupStatefulSet(testNamespace1); err != nil {
		t.Fatalf("Failed to setup StatefulSet: %v", err)
	}
	defer teardownStatefulSet(testNamespace1)

	ctx := context.Background()

	labels, err := getStatefulSetSelectorLabels(ctx, testStatefulSet, testNamespace1)
	if err != nil {
		t.Fatalf("Failed to get selector labels: %v", err)
	}

	if labels == nil {
		t.Fatal("Expected to get selector labels, got nil")
	}

	if _, ok := labels["app"]; !ok {
		t.Errorf("Expected 'app' label in selector, got: %v", labels)
	}

	if labels["app"] != testStatefulSet {
		t.Errorf("Expected app=%s, got app=%s", testStatefulSet, labels["app"])
	}

	t.Logf("Got StatefulSet selector labels: %v", labels)
}
