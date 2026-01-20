package portforward

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GiGurra/cmder"
)

// SourceKind represents the type of Kubernetes resource to port-forward from
type SourceKind string

const (
	SourceDeployment  SourceKind = "deployment"
	SourceStatefulSet SourceKind = "statefulset"
	SourceDaemonSet   SourceKind = "daemonset"
	SourceService     SourceKind = "service"
)

// Source represents a Kubernetes resource that can be used as a port-forward target
type Source struct {
	Kind SourceKind
	Name string
}

// String returns a human-readable description of the source
func (s Source) String() string {
	return fmt.Sprintf("%s/%s", s.Kind, s.Name)
}

// IsEmpty returns true if no source is configured
func (s Source) IsEmpty() bool {
	return s.Name == ""
}

// GetSelectorLabels returns the selector labels for finding pods from this source
func (s Source) GetSelectorLabels(ctx context.Context, namespace string) (map[string]string, error) {
	switch s.Kind {
	case SourceDeployment:
		return getDeploymentSelectorLabels(ctx, s.Name, namespace)
	case SourceStatefulSet:
		return getStatefulSetSelectorLabels(ctx, s.Name, namespace)
	case SourceDaemonSet:
		return getDaemonSetSelectorLabels(ctx, s.Name, namespace)
	case SourceService:
		return getServiceSelectorLabels(ctx, s.Name, namespace)
	default:
		return nil, fmt.Errorf("unknown source kind: %s", s.Kind)
	}
}

// findResources is a generic function to list resources of a given kind
func findResources(ctx context.Context, kind string, namespace string) []string {
	cmdAndArgs := []string{"kubectl", "get", kind, "-o", "name"}
	if namespace != "" {
		cmdAndArgs = append(cmdAndArgs, "-n", namespace)
	}

	res := cmder.New(cmdAndArgs...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)
	if res.Err != nil {
		return nil
	}

	var resources []string
	lines := strings.Split(strings.TrimSpace(res.StdOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove the "kind/" prefix (e.g., "deployment/", "service/", etc.)
		parts := strings.SplitN(line, "/", 2)
		if len(parts) == 2 {
			resources = append(resources, parts[1])
		}
	}
	return resources
}
