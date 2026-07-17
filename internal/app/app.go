package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"git-muse/internal/commitmsg"
	"git-muse/internal/diffctx"
	"git-muse/internal/git"
)

func RunInteractive(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}

	ctxData, err := diffctx.Collect(ctx, client)
	if err != nil {
		return err
	}

	msg := commitmsg.Render(commitmsg.Candidate{
		Type:    "chore",
		Subject: "describe staged changes",
	})

	fmt.Println("Suggested commit message:")
	fmt.Println()
	fmt.Print(msg)
	fmt.Println()
	fmt.Println("LLM provider not wired yet. Use `git muse inspect` to review collected context.")
	fmt.Fprintf(os.Stderr, "staged diff bytes: %d\n", len(ctxData.Diff))
	return nil
}

func RunGenerate(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}

	if _, err := diffctx.Collect(ctx, client); err != nil {
		return err
	}

	fmt.Print(commitmsg.Render(commitmsg.Candidate{
		Type:    "chore",
		Subject: "describe staged changes",
	}))
	return nil
}

func RunInspect(ctx context.Context) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}

	data, err := diffctx.Collect(ctx, client)
	if err != nil {
		return err
	}

	fmt.Printf("Branch: %s\n", data.Branch)
	fmt.Printf("Status bytes: %d\n", len(data.Status))
	fmt.Printf("Stat bytes: %d\n", len(data.Stat))
	fmt.Printf("Diff bytes: %d\n", len(data.Diff))
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

	data, err := diffctx.Collect(ctx, client)
	if err != nil {
		return err
	}

	fmt.Println("Git repository      OK")
	fmt.Printf("Staged changes      OK, %d diff bytes\n", len(data.Diff))
	fmt.Println("Provider            not configured")
	fmt.Println("Model               not configured")
	return nil
}
