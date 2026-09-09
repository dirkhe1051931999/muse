package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"git-muse/internal/app"
	"git-muse/internal/docs"
	"git-muse/internal/version"
)

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
		printHelp()
		return nil
	}
	if flags.Version {
		fmt.Print(version.Identity())
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
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", flags.Rest[0], docs.Usage)
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

func printHelp() {
	fmt.Print(docs.Usage)
	fmt.Print("\n")
	fmt.Print(version.Identity())
}
