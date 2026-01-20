package portforward

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GiGurra/cmder"
	"github.com/samber/lo"
)

// getDaemonSetSelectorLabels returns the selector labels for a daemonset
func getDaemonSetSelectorLabels(ctx context.Context, name string, namespace string) (map[string]string, error) {
	found := lo.Filter(findDaemonSets(ctx, namespace), func(item string, _ int) bool {
		return item == name
	})

	if len(found) == 0 {
		return nil, fmt.Errorf("daemonset %s not found", name)
	}

	args := []string{"get", "daemonset", name, "-o", "jsonpath={.spec.selector.matchLabels}"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return nil, fmt.Errorf("failed to get daemonset %s selector labels: %w", name, result.Err)
	}

	labels := make(map[string]string)
	if err := json.Unmarshal([]byte(result.StdOut), &labels); err != nil {
		return nil, fmt.Errorf("failed to parse daemonset %s selector labels: %w", name, err)
	}

	return labels, nil
}

// findDaemonSets returns a list of daemonset names in the given namespace
func findDaemonSets(ctx context.Context, namespace string) []string {
	return findResources(ctx, "daemonsets", namespace)
}
