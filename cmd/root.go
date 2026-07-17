package cmd

import (
	"context"
	"fmt"
	"os"

	"git-muse/internal/app"
)

const usage = `git-muse - AI commit message helper

Usage:
  git muse [command]
  git-muse [command]

Commands:
  generate   Generate commit message from staged changes
  inspect    Show context that would be sent to LLM
  doctor     Check repository and configuration
  help       Show help
`

func Execute() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return app.RunInteractive(ctx)
	}

	switch args[0] {
	case "generate":
		return app.RunGenerate(ctx)
	case "inspect":
		return app.RunInspect(ctx)
	case "doctor":
		return app.RunDoctor(ctx)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}
