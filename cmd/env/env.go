package env

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Command []string `pos:"true" optional:"true" help:"Command to run with modified environment."`
	Format  string   `short:"f" help:"Output format (plain, json, shell, powershell)." default:"plain" alts:"plain,json,shell,powershell"`
	Filter  string   `help:"Filter variables by prefix (case-insensitive)." optional:"true"`
	Sort    bool     `short:"s" help:"Sort variables alphabetically." default:"true"`
	Keys    bool     `short:"k" help:"Show only variable names (keys)." optional:"true"`
	Values  bool     `short:"v" help:"Show only variable values." optional:"true"`
	Get     string   `short:"g" help:"Get a specific environment variable." optional:"true"`
	Set     string   `help:"Set an environment variable (format: KEY=VALUE) and run command." optional:"true"`
	Unset   string   `short:"u" help:"Unset an environment variable and run command." optional:"true"`
	Export  bool     `short:"e" help:"Output in export format for shell sourcing." optional:"true"`
	NoEmpty bool     `help:"Hide variables with empty values." optional:"true"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "env",
		Short:       "Cross-platform environment variable management",
		Long:        "List, get, set, or filter environment variables. Works consistently across Windows, macOS, and Linux.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runEnv(params); err != nil {
				fmt.Fprintf(os.Stderr, "env: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runEnv(params *Params) error {
	// Handle --get: retrieve a single variable
	if params.Get != "" {
		value, exists := os.LookupEnv(params.Get)
		if !exists {
			return fmt.Errorf("environment variable %q not set", params.Get)
		}
		fmt.Println(value)
		return nil
	}

	// Handle --set: set a variable and optionally run a command
	if params.Set != "" {
		parts := strings.SplitN(params.Set, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format for --set, expected KEY=VALUE")
		}
		key, value := parts[0], parts[1]
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}

		// If command provided, run it with modified environment
		if len(params.Command) > 0 {
			return runWithEnv(params.Command)
		}

		// Otherwise just confirm the set
		fmt.Printf("%s=%s\n", key, value)
		return nil
	}

	// Handle --unset: unset a variable and optionally run a command
	if params.Unset != "" {
		if err := os.Unsetenv(params.Unset); err != nil {
			return fmt.Errorf("failed to unset %s: %w", params.Unset, err)
		}

		// If command provided, run it with modified environment
		if len(params.Command) > 0 {
			return runWithEnv(params.Command)
		}

		fmt.Printf("unset %s\n", params.Unset)
		return nil
	}

	// If command is provided without set/unset, run it with current env
	if len(params.Command) > 0 {
		return runWithEnv(params.Command)
	}

	// Default: list environment variables
	return listEnv(params)
}

func runWithEnv(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func listEnv(params *Params) error {
	envVars := os.Environ()
	envMap := make(map[string]string)
	var keys []string

	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := ""
		if len(parts) > 1 {
			value = parts[1]
		}

		// Apply filter
		if params.Filter != "" {
			if !strings.HasPrefix(strings.ToUpper(key), strings.ToUpper(params.Filter)) {
				continue
			}
		}

		// Skip empty values if requested
		if params.NoEmpty && value == "" {
			continue
		}

		envMap[key] = value
		keys = append(keys, key)
	}

	// Sort if requested
	if params.Sort {
		sort.Strings(keys)
	}

	// Handle export format override
	format := params.Format
	if params.Export {
		if runtime.GOOS == "windows" {
			format = "powershell"
		} else {
			format = "shell"
		}
	}

	// Output based on format
	switch format {
	case "json":
		return outputJSON(envMap, keys, params)
	case "shell":
		return outputShell(envMap, keys, params)
	case "powershell":
		return outputPowershell(envMap, keys, params)
	default:
		return outputPlain(envMap, keys, params)
	}
}

func outputPlain(envMap map[string]string, keys []string, params *Params) error {
	for _, key := range keys {
		if params.Keys && !params.Values {
			fmt.Println(key)
		} else if params.Values && !params.Keys {
			fmt.Println(envMap[key])
		} else {
			fmt.Printf("%s=%s\n", key, envMap[key])
		}
	}
	return nil
}

func outputJSON(envMap map[string]string, keys []string, params *Params) error {
	var output interface{}

	if params.Keys && !params.Values {
		output = keys
	} else if params.Values && !params.Keys {
		values := make([]string, len(keys))
		for i, key := range keys {
			values[i] = envMap[key]
		}
		output = values
	} else {
		// Preserve order by creating ordered map representation
		orderedMap := make(map[string]string)
		for _, key := range keys {
			orderedMap[key] = envMap[key]
		}
		output = orderedMap
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputShell(envMap map[string]string, keys []string, params *Params) error {
	for _, key := range keys {
		if params.Keys && !params.Values {
			fmt.Println(key)
		} else if params.Values && !params.Keys {
			fmt.Printf("%q\n", envMap[key])
		} else {
			// Escape for shell
			value := strings.ReplaceAll(envMap[key], "'", "'\"'\"'")
			fmt.Printf("export %s='%s'\n", key, value)
		}
	}
	return nil
}

func outputPowershell(envMap map[string]string, keys []string, params *Params) error {
	for _, key := range keys {
		if params.Keys && !params.Values {
			fmt.Println(key)
		} else if params.Values && !params.Keys {
			fmt.Printf("%q\n", envMap[key])
		} else {
			// Escape for PowerShell
			value := strings.ReplaceAll(envMap[key], "'", "''")
			fmt.Printf("$env:%s = '%s'\n", key, value)
		}
	}
	return nil
}
