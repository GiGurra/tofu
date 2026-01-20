package portforward

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GiGurra/cmder"
	"github.com/samber/lo"
)

// getStatefulSetSelectorLabels returns the selector labels for a statefulset
func getStatefulSetSelectorLabels(ctx context.Context, name string, namespace string) (map[string]string, error) {
	found := lo.Filter(findStatefulSets(ctx, namespace), func(item string, _ int) bool {
		return item == name
	})

	if len(found) == 0 {
		return nil, fmt.Errorf("statefulset %s not found", name)
	}

	args := []string{"get", "statefulset", name, "-o", "jsonpath={.spec.selector.matchLabels}"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return nil, fmt.Errorf("failed to get statefulset %s selector labels: %w", name, result.Err)
	}

	labels := make(map[string]string)
	if err := json.Unmarshal([]byte(result.StdOut), &labels); err != nil {
		return nil, fmt.Errorf("failed to parse statefulset %s selector labels: %w", name, err)
	}

	return labels, nil
}

// findStatefulSets returns a list of statefulset names in the given namespace
func findStatefulSets(ctx context.Context, namespace string) []string {
	return findResources(ctx, "statefulsets", namespace)
}
