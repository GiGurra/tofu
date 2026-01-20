package portforward

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GiGurra/cmder"
	"github.com/samber/lo"
)

// getServiceSelectorLabels returns the selector labels for a service
// Note: Services use .spec.selector directly, not .spec.selector.matchLabels
func getServiceSelectorLabels(ctx context.Context, name string, namespace string) (map[string]string, error) {
	found := lo.Filter(findServices(ctx, namespace), func(item string, _ int) bool {
		return item == name
	})

	if len(found) == 0 {
		return nil, fmt.Errorf("service %s not found", name)
	}

	args := []string{"get", "service", name, "-o", "jsonpath={.spec.selector}"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	result := cmder.New(append([]string{"kubectl"}, args...)...).
		WithAttemptTimeout(5 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return nil, fmt.Errorf("failed to get service %s selector: %w", name, result.Err)
	}

	if result.StdOut == "" {
		return nil, fmt.Errorf("service %s has no selector (possibly an ExternalName service)", name)
	}

	labels := make(map[string]string)
	if err := json.Unmarshal([]byte(result.StdOut), &labels); err != nil {
		return nil, fmt.Errorf("failed to parse service %s selector: %w", name, err)
	}

	if len(labels) == 0 {
		return nil, fmt.Errorf("service %s has empty selector", name)
	}

	return labels, nil
}

// findServices returns a list of service names in the given namespace
func findServices(ctx context.Context, namespace string) []string {
	return findResources(ctx, "services", namespace)
}
