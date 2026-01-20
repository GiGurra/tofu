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
	testService = "nginx-svc-test"
)

func setupService(ns string) error {
	ctx := context.Background()

	// Create a Service that selects the existing deployment pods
	// The deployment (testDeployment = "nginx-test") is created in the main test setup
	svcYaml := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  selector:
    app: %s
  ports:
  - port: 80
    targetPort: 80
`, testService, ns, testDeployment)

	result := cmder.New("kubectl", "apply", "-f", "-").
		WithStdIn(strings.NewReader(svcYaml)).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return fmt.Errorf("failed to create service: %w\n%s", result.Err, result.Combined)
	}

	return nil
}

func teardownService(ns string) {
	ctx := context.Background()
	_ = cmder.New("kubectl", "delete", "service", testService, "-n", ns, "--ignore-not-found").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
}

func TestPortForward_Service(t *testing.T) {
	// Setup Service in test namespace (uses existing deployment pods)
	if err := setupService(testNamespace1); err != nil {
		t.Fatalf("Failed to setup Service: %v", err)
	}
	defer teardownService(testNamespace1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		Source:    Source{Kind: SourceService, Name: testService},
		Namespace: testNamespace1,
		Ports:     []string{"18092:80"},
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
	resp, err := http.Get("http://localhost:18092")
	if err != nil {
		t.Fatalf("Failed to connect to nginx via Service port-forward: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	t.Log("Successfully connected to nginx via Service port-forward")
	cancel()

	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
	}
}

func TestGetServiceSelectorLabels(t *testing.T) {
	// Setup Service
	if err := setupService(testNamespace1); err != nil {
		t.Fatalf("Failed to setup Service: %v", err)
	}
	defer teardownService(testNamespace1)

	ctx := context.Background()

	labels, err := getServiceSelectorLabels(ctx, testService, testNamespace1)
	if err != nil {
		t.Fatalf("Failed to get selector labels: %v", err)
	}

	if labels == nil {
		t.Fatal("Expected to get selector labels, got nil")
	}

	if _, ok := labels["app"]; !ok {
		t.Errorf("Expected 'app' label in selector, got: %v", labels)
	}

	// Service selects the deployment's pods, so app should be testDeployment
	if labels["app"] != testDeployment {
		t.Errorf("Expected app=%s, got app=%s", testDeployment, labels["app"])
	}

	t.Logf("Got Service selector labels: %v", labels)
}

func TestPortForward_ServiceNotFound(t *testing.T) {
	cfg := Config{
		Source:    Source{Kind: SourceService, Name: "nonexistent-service"},
		Namespace: testNamespace1,
		Ports:     []string{"18093:80"},
		Keepalive: false,
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
			t.Error("Expected error for nonexistent service, got nil")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	case <-ctx.Done():
		t.Log("Port-forward correctly failed to find service")
	}
}
