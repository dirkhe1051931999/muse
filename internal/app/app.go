package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"git-muse/internal/commitmsg"
	"git-muse/internal/config"
	"git-muse/internal/diffctx"
	"git-muse/internal/git"
	"git-muse/internal/llm"
	"git-muse/internal/ui"
)

func RunInteractive(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg := config.Load(ctx, client)
	if err := cfg.ValidateProvider(); err != nil {
		return err
	}

	ctxData, err := diffctx.Collect(ctx, client, cfg.MaxDiffBytes)
	if err != nil {
		return err
	}

	provider, err := llm.New(cfg)
	if err != nil {
		return err
	}

	generate := func() (string, error) {
		msg, err := generateMessage(ctx, provider, cfg, ctxData)
		if err != nil {
			return "", err
		}
		return msg, nil
	}

	msg, err := generate()
	if err != nil {
		return err
	}
	final, ok, err := ui.Choose(ctx, client, msg, ctxData.Diff, generate)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}
	if !ui.Confirm(final) {
		fmt.Println("Cancelled.")
		return nil
	}
	return commitWithMessage(ctx, client, final)
}

func RunGenerate(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg := config.Load(ctx, client)
	if err := cfg.ValidateProvider(); err != nil {
		return err
	}

	data, err := diffctx.Collect(ctx, client, cfg.MaxDiffBytes)
	if err != nil {
		return err
	}
	provider, err := llm.New(cfg)
	if err != nil {
		return err
	}
	msg, err := generateMessage(ctx, provider, cfg, data)
	if err != nil {
		return err
	}
	fmt.Print(msg)
	return nil
}

func RunInspect(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg := config.Load(ctx, client)

	data, err := diffctx.Collect(ctx, client, cfg.MaxDiffBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Provider: %s\n", valueOrUnset(cfg.Provider))
	fmt.Printf("Model: %s\n", valueOrUnset(cfg.Model))
	fmt.Printf("Base URL: %s\n", valueOrUnset(cfg.BaseURL))
	fmt.Printf("Branch: %s\n", data.Branch)
	fmt.Printf("Status bytes: %d\n", len(data.Status))
	fmt.Printf("Stat bytes: %d\n", len(data.Stat))
	fmt.Printf("Included diff bytes: %d\n", len(data.Diff))
	fmt.Printf("Truncated: %v\n", data.Truncated)
	if len(data.Excluded) > 0 {
		fmt.Println("Excluded:")
		for _, item := range data.Excluded {
			fmt.Printf("  %s  %s\n", item.Path, item.Reason)
		}
	}
	if len(data.RecentSubjects) > 0 {
		fmt.Println("Recent commits:")
		fmt.Println(strings.Join(data.RecentSubjects, "\n"))
	}
	return nil
}

func RunDoctor(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg := config.Load(ctx, client)

	data, err := diffctx.Collect(ctx, client, cfg.MaxDiffBytes)
	if err != nil {
		return err
	}

	fmt.Println("Git repository      OK")
	fmt.Printf("Staged changes      OK, %d diff bytes\n", len(data.Diff))
	fmt.Printf("Provider            %s\n", statusValue(cfg.Provider))
	fmt.Printf("Model               %s\n", statusValue(cfg.Model))
	if cfg.Provider == "ollama" {
		fmt.Println("API key             not required")
	} else if cfg.APIKeyName != "" && cfg.APIKey != "" {
		fmt.Printf("API key             OK, %s\n", cfg.APIKeyName)
	} else if cfg.APIKeyName != "" {
		fmt.Printf("API key             missing, %s\n", cfg.APIKeyName)
	} else {
		fmt.Println("API key             not checked")
	}
	return nil
}

func generateMessage(ctx context.Context, provider llm.Provider, cfg config.Config, data diffctx.Context) (string, error) {
	response, err := provider.Generate(ctx, llm.GenerateRequest{Config: cfg, Diff: data})
	if err != nil {
		response, err = provider.Generate(ctx, llm.GenerateRequest{Config: cfg, Diff: data})
		if err != nil {
			return "", err
		}
	}
	return commitmsg.Render(response.Candidates[0]), nil
}

func commitWithMessage(ctx context.Context, client git.Client, message string) error {
	file, err := os.CreateTemp("", "git-muse-commit-*.txt")
	if err != nil {
		return err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.WriteString(message); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return client.CommitWithFile(ctx, path)
}

func valueOrUnset(value string) string {
	if value == "" {
		return "not configured"
	}
	return value
}

func statusValue(value string) string {
	if value == "" {
		return "not configured"
	}
	return value
}
