package listrepos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type Params struct {
	Org  boa.Required[string] `help:"GitHub organization name"`
	Team boa.Optional[string] `help:"Team slug/name within the organization (optional, lists all org repos if not specified)"`
	Url  bool                 `help:"Print full GitHub URLs instead of repo names" optional:"true"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "list-repos",
		Short:       "List repositories for an org or team",
		Long:        "List all repositories in an organization, optionally filtered by team. Requires gh CLI to be installed and authenticated.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			params.Org.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				orgs := listOrgs(context.Background())
				return lo.Filter(orgs, func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			params.Team.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				if params.Org.Value() == "" {
					return nil
				}
				teams := listTeams(context.Background(), params.Org.Value())
				return lo.Filter(teams, func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				})
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

func run(params *Params, stdout, _ io.Writer) error {
	if err := checkGh(); err != nil {
		return err
	}

	var repos []string
	var err error

	if params.Team.HasValue() {
		repos, err = listTeamRepos(context.Background(), params.Org.Value(), *params.Team.Value())
	} else {
		repos, err = listOrgRepos(context.Background(), params.Org.Value())
	}
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if params.Url {
			fmt.Fprintf(stdout, "https://github.com/%s\n", repo)
		} else {
			fmt.Fprintln(stdout, repo)
		}
	}

	return nil
}

func checkGh() error {
	result := cmder.New("gh", "version").
		WithAttemptTimeout(5 * time.Second).
		Run(context.Background())
	if result.Err != nil {
		if result.Combined != "" {
			return fmt.Errorf("gh CLI not found or not working: %w\n%s", result.Err, result.Combined)
		}
		return fmt.Errorf("gh CLI not found or not working: %w", result.Err)
	}
	return nil
}

func listOrgs(ctx context.Context) []string {
	result := cmder.New("gh", "api", "/user/orgs", "--jq", ".[].login").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return nil
	}

	var orgs []string
	lines := strings.Split(strings.TrimSpace(result.StdOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			orgs = append(orgs, line)
		}
	}
	return orgs
}

func listTeams(ctx context.Context, org string) []string {
	result := cmder.New("gh", "api", fmt.Sprintf("/orgs/%s/teams", org), "--jq", ".[].slug").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return nil
	}

	var teams []string
	lines := strings.Split(strings.TrimSpace(result.StdOut), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			teams = append(teams, line)
		}
	}
	return teams
}

type repoResponse struct {
	FullName string `json:"full_name"`
}

func listTeamRepos(ctx context.Context, org, team string) ([]string, error) {
	result := cmder.New("gh", "api", fmt.Sprintf("/orgs/%s/teams/%s/repos", org, team), "--paginate").
		WithAttemptTimeout(60 * time.Second).
		Run(ctx)
	if result.Err != nil {
		if result.Combined != "" {
			return nil, fmt.Errorf("failed to list team repos: %w\n%s", result.Err, result.Combined)
		}
		return nil, fmt.Errorf("failed to list team repos: %w", result.Err)
	}

	var repos []repoResponse
	if err := json.Unmarshal([]byte(result.StdOut), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return lo.Map(repos, func(r repoResponse, _ int) string {
		return r.FullName
	}), nil
}

func listOrgRepos(ctx context.Context, org string) ([]string, error) {
	result := cmder.New("gh", "api", fmt.Sprintf("/orgs/%s/repos", org), "--paginate").
		WithAttemptTimeout(60 * time.Second).
		Run(ctx)
	if result.Err != nil {
		if result.Combined != "" {
			return nil, fmt.Errorf("failed to list org repos: %w\n%s", result.Err, result.Combined)
		}
		return nil, fmt.Errorf("failed to list org repos: %w", result.Err)
	}

	var repos []repoResponse
	if err := json.Unmarshal([]byte(result.StdOut), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return lo.Map(repos, func(r repoResponse, _ int) string {
		return r.FullName
	}), nil
}
