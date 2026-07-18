package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"git-muse/internal/app"
	"git-muse/internal/version"
)

const usage = `git-muse - AI commit message helper

Usage:
  git muse [command]
  git-muse [command]
  git muse --help
  git muse --version
  git muse --proxy http://127.0.0.1:7890
  git muse --proxy 127.0.0.1:7890

Commands:
  init       Write example config to ~/.git-muse.json
  generate   Generate commit message from staged changes
  inspect    Show context that would be sent to LLM
  doctor     Check repository and configuration
  help       Show help

Global flags:
  -h, --help              Show help
  -V, --version           Show version
      --proxy <url>       HTTP proxy for LLM requests (http/https targets)
`

type globalFlags struct {
	Help    bool
	Version bool
	Proxy   string
	Rest    []string
}

func Execute() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags, err := parseGlobalFlags(args)
	if err != nil {
		return err
	}
	if flags.Help {
		fmt.Print(usage)
		return nil
	}
	if flags.Version {
		fmt.Printf("git-muse %s\n", version.Version)
		return nil
	}

	opts := app.Options{Proxy: flags.Proxy}
	if len(flags.Rest) == 0 {
		return app.RunInteractive(ctx, opts)
	}

	switch flags.Rest[0] {
	case "init":
		return app.RunInit(ctx)
	case "generate":
		return app.RunGenerate(ctx, opts)
	case "inspect":
		return app.RunInspect(ctx, opts)
	case "doctor":
		return app.RunDoctor(ctx, opts)
	case "help":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", flags.Rest[0], usage)
	}
}

func parseGlobalFlags(args []string) (globalFlags, error) {
	var flags globalFlags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			flags.Help = true
		case arg == "--version" || arg == "-V":
			flags.Version = true
		case arg == "--proxy":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--proxy requires a value\n\nExample:\n\n  git muse --proxy http://127.0.0.1:7890")
			}
			i++
			flags.Proxy = args[i]
		case strings.HasPrefix(arg, "--proxy="):
			flags.Proxy = strings.TrimPrefix(arg, "--proxy=")
		default:
			flags.Rest = append(flags.Rest, arg)
		}
	}
	return flags, nil
}
