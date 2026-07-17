package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	Dir string
}

func New(dir string) Client {
	return Client{Dir: dir}
}

func (c Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.Dir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}

func (c Client) RunStdin(ctx context.Context, stdin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.Dir
	cmd.Stdin = strings.NewReader(stdin)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}

func (c Client) RequireRepo(ctx context.Context) error {
	out, err := c.Run(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("not inside Git repository")
	}
	if strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf("not inside Git repository")
	}
	return nil
}

func (c Client) StagedDiff(ctx context.Context) (string, error) {
	out, err := c.Run(ctx, "diff", "--cached", "--no-ext-diff", "--unified=3")
	return string(out), err
}

func (c Client) StagedStat(ctx context.Context) (string, error) {
	out, err := c.Run(ctx, "diff", "--cached", "--stat")
	return string(out), err
}

func (c Client) StatusPorcelain(ctx context.Context) (string, error) {
	out, err := c.Run(ctx, "status", "--porcelain=v1")
	return string(out), err
}

func (c Client) BranchName(ctx context.Context) (string, error) {
	out, err := c.Run(ctx, "branch", "--show-current")
	return strings.TrimSpace(string(out)), err
}

func (c Client) RecentSubjects(ctx context.Context, n int) ([]string, error) {
	out, err := c.Run(ctx, "log", fmt.Sprintf("-%d", n), "--pretty=format:%s")
	if err != nil {
		return nil, nil
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

func (c Client) Config(ctx context.Context, key string) (string, error) {
	out, err := c.Run(ctx, "config", "--get", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (c Client) Editor(ctx context.Context) string {
	for _, key := range []string{"GIT_EDITOR", "VISUAL", "EDITOR"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	out, err := c.Run(ctx, "var", "GIT_EDITOR")
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func (c Client) CommitWithFile(ctx context.Context, path string) error {
	_, err := c.Run(ctx, "commit", "-F", path)
	return err
}
