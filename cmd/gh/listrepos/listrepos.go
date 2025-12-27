package listrepos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/GiGurra/cmder"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type Params struct {
	Owner      boa.Required[string] `short:"o" help:"GitHub organization or user name"`
	Team       boa.Optional[string] `help:"Team slug/name within the organization (optional, lists all repos if not specified)"`
	Visibility []string             `short:"v" help:"Filter by visibility (can specify multiple)" default:"all" alts:"all,public,private,internal"`
	Archived   string               `help:"Filter by archived status" default:"all" alts:"all,archived,not-archived"`
	Sort       string               `help:"Sort repos by field" default:"full_name" alts:"full_name,created,updated,pushed"`
	Direction  string               `help:"Sort direction" default:"asc" alts:"asc,desc"`
	Url        bool                 `help:"Print full GitHub URLs instead of repo names" optional:"true"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "list-repos",
		Short:       "List repositories for a user, org, or team",
		Long:        "List all repositories for a GitHub user or organization, optionally filtered by team. Requires gh CLI to be installed and authenticated.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			params.Owner.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				orgs := listOrgs(context.Background())
				return lo.Filter(orgs, func(item string, index int) bool {
					return strings.HasPrefix(item, toComplete)
				})
			}

			params.Team.AlternativesFunc = func(cmd *cobra.Command, args []string, toComplete string) []string {
				if params.Owner.Value() == "" {
					return nil
				}
				teams := listTeams(context.Background(), params.Owner.Value())
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

	var repos []repoResponse
	var err error

	if params.Team.HasValue() {
		repos, err = listTeamRepos(context.Background(), params.Owner.Value(), *params.Team.Value())
	} else {
		repos, err = listOwnerRepos(context.Background(), params.Owner.Value())
	}
	if err != nil {
		return err
	}

	// Apply filters
	filter := repoFilter{
		visibility: params.Visibility,
		archived:   params.Archived,
	}
	repos = lo.Filter(repos, func(r repoResponse, _ int) bool {
		return filter.matches(r)
	})

	// Apply sorting
	sortRepos(repos, params.Sort, params.Direction)

	for _, repo := range repos {
		if params.Url {
			fmt.Fprintf(stdout, "https://github.com/%s\n", repo.FullName)
		} else {
			fmt.Fprintln(stdout, repo.FullName)
		}
	}

	return nil
}

func sortRepos(repos []repoResponse, sortBy, direction string) {
	slices.SortFunc(repos, func(a, b repoResponse) int {
		var cmp int
		switch sortBy {
		case "created":
			cmp = strings.Compare(a.CreatedAt, b.CreatedAt)
		case "updated":
			cmp = strings.Compare(a.UpdatedAt, b.UpdatedAt)
		case "pushed":
			cmp = strings.Compare(a.PushedAt, b.PushedAt)
		default: // full_name
			cmp = strings.Compare(strings.ToLower(a.FullName), strings.ToLower(b.FullName))
		}
		if direction == "desc" {
			cmp = -cmp
		}
		return cmp
	})
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
	FullName   string `json:"full_name"`
	Visibility string `json:"visibility"`
	Archived   bool   `json:"archived"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	PushedAt   string `json:"pushed_at"`
}

type repoFilter struct {
	visibility []string // empty means all
	archived   string   // "all", "archived", "not-archived"
}

func (f repoFilter) matches(repo repoResponse) bool {
	// Check visibility
	if len(f.visibility) > 0 && !lo.Contains(f.visibility, "all") {
		if !lo.Contains(f.visibility, repo.Visibility) {
			return false
		}
	}

	// Check archived status
	switch f.archived {
	case "archived":
		if !repo.Archived {
			return false
		}
	case "not-archived":
		if repo.Archived {
			return false
		}
	}

	return true
}

func listTeamRepos(ctx context.Context, org, team string) ([]repoResponse, error) {
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

	return repos, nil
}

func listOwnerRepos(ctx context.Context, owner string) ([]repoResponse, error) {
	// Check if owner is a user or org
	isOrg, err := isOrganization(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to determine account type: %w", err)
	}

	var endpoint string
	if isOrg {
		endpoint = fmt.Sprintf("/orgs/%s/repos", owner)
	} else {
		// Check if owner is the authenticated user
		// /users/{username}/repos only returns public repos
		// /user/repos returns all repos (including private) for the authenticated user
		if isAuthenticatedUser(ctx, owner) {
			endpoint = "/user/repos"
		} else {
			endpoint = fmt.Sprintf("/users/%s/repos", owner)
		}
	}

	result := cmder.New("gh", "api", endpoint, "--paginate").
		WithAttemptTimeout(60 * time.Second).
		Run(ctx)

	if result.Err != nil {
		if result.Combined != "" {
			return nil, fmt.Errorf("failed to list repos: %w\n%s", result.Err, result.Combined)
		}
		return nil, fmt.Errorf("failed to list repos: %w", result.Err)
	}

	var repos []repoResponse
	if err := json.Unmarshal([]byte(result.StdOut), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return repos, nil
}

func isAuthenticatedUser(ctx context.Context, owner string) bool {
	result := cmder.New("gh", "api", "/user", "--jq", ".login").
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)
	if result.Err != nil {
		return false
	}
	return strings.TrimSpace(result.StdOut) == owner
}

type ownerResponse struct {
	Type string `json:"type"`
}

func isOrganization(ctx context.Context, owner string) (bool, error) {
	result := cmder.New("gh", "api", fmt.Sprintf("/users/%s", owner)).
		WithAttemptTimeout(10 * time.Second).
		Run(ctx)

	if result.Err != nil {
		if result.Combined != "" {
			return false, fmt.Errorf("%w\n%s", result.Err, result.Combined)
		}
		return false, result.Err
	}

	var resp ownerResponse
	if err := json.Unmarshal([]byte(result.StdOut), &resp); err != nil {
		return false, err
	}

	return resp.Type == "Organization", nil
}
