package diffctx

import (
	"context"
	"fmt"
	"strings"

	"git-muse/internal/git"
)

type Context struct {
	Branch         string
	RecentSubjects []string
	Status         string
	Stat           string
	Diff           string
	Excluded       []ExcludedFile
	Truncated      bool
}

type ExcludedFile struct {
	Path   string
	Reason string
}

func Collect(ctx context.Context, client git.Client, opts FilterOptions) (Context, error) {
	diff, err := client.StagedDiff(ctx)
	if err != nil {
		return Context{}, err
	}
	if strings.TrimSpace(diff) == "" {
		return Context{}, fmt.Errorf("no staged changes found\n\nStage files first:\n\n  git add <files>")
	}

	filteredDiff, excluded, truncated := Filter(diff, opts)

	stat, err := client.StagedStat(ctx)
	if err != nil {
		return Context{}, err
	}
	status, err := client.StatusPorcelain(ctx)
	if err != nil {
		return Context{}, err
	}
	branch, _ := client.BranchName(ctx)
	recent, _ := client.RecentSubjects(ctx, 5)

	return Context{
		Branch:         branch,
		RecentSubjects: recent,
		Status:         status,
		Stat:           stat,
		Diff:           filteredDiff,
		Excluded:       excluded,
		Truncated:      truncated,
	}, nil
}
