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
}

func Collect(ctx context.Context, client git.Client) (Context, error) {
	diff, err := client.StagedDiff(ctx)
	if err != nil {
		return Context{}, err
	}
	if strings.TrimSpace(diff) == "" {
		return Context{}, fmt.Errorf("no staged changes found\n\nStage files first:\n\n  git add <files>")
	}

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
		Diff:           diff,
	}, nil
}
