package app

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git-muse/internal/commitmsg"
	"git-muse/internal/config"
	"git-muse/internal/diffctx"
	"git-muse/internal/docs"
	"git-muse/internal/git"
	"git-muse/internal/llm"
	"git-muse/internal/ui"
	"git-muse/internal/version"
)

type Options struct {
	Proxy string
}

func RunInteractive(ctx context.Context, opts Options) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg, err := loadConfig(ctx, client, opts)
	if err != nil {
		return err
	}
	if err := cfg.ValidateProvider(); err != nil {
		return err
	}

	ctxData, err := diffctx.Collect(ctx, client, filterOptions(cfg))
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
	final, ok, push, err := ui.Choose(ctx, client, msg, generate)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}
	if !ui.Confirm(final, push) {
		fmt.Println("Cancelled.")
		return nil
	}
	if err := commitWithMessage(ctx, client, final); err != nil {
		return err
	}
	if push {
		return client.Push(ctx)
	}
	return nil
}

func RunGenerate(ctx context.Context, opts Options) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg, err := loadConfig(ctx, client, opts)
	if err != nil {
		return err
	}
	if err := cfg.ValidateProvider(); err != nil {
		return err
	}

	data, err := diffctx.Collect(ctx, client, filterOptions(cfg))
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

func RunInspect(ctx context.Context, opts Options) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg, err := loadConfig(ctx, client, opts)
	if err != nil {
		return err
	}
	if cfg.FileError != nil {
		return cfg.FileError
	}

	data, err := diffctx.Collect(ctx, client, filterOptions(cfg))
	if err != nil {
		return err
	}

	printLLMEndpoints(cfg)
	fmt.Printf("Proxy: %s\n", valueOrUnset(cfg.Proxy))
	fmt.Printf("Language: %s\n", cfg.Language)
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

func RunInit(ctx context.Context) error {
	path, err := config.Path()
	if err != nil {
		return err
	}
	overwrite := true
	if _, err := os.Stat(path); err == nil {
		overwrite = confirmOverwrite(path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if overwrite {
		path, err = config.InitFile(true)
		if err != nil {
			return err
		}
		fmt.Printf("Wrote %s\n", path)
		fmt.Println("Edit apiKey before running git muse.")
	} else {
		fmt.Printf("Skipped %s\n", path)
	}
	if err := addExeDirToUserPath(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "PATH update skipped: %v\n", err)
	}
	if htmlPath, err := docs.InstallGitHTML(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Git help page skipped: %v\n", err)
		fmt.Fprintln(os.Stderr, "Use: git muse help   or   git-muse --help")
	} else {
		fmt.Printf("Installed Git help page %s\n", htmlPath)
	}
	return nil
}

func RunDoctor(ctx context.Context, opts Options) error {
	client := git.New("")
	if err := client.RequireRepo(ctx); err != nil {
		return err
	}
	cfg, err := loadConfig(ctx, client, opts)
	if err != nil {
		return err
	}
	if cfg.FileError != nil {
		return cfg.FileError
	}

	data, err := diffctx.Collect(ctx, client, filterOptions(cfg))
	if err != nil {
		return err
	}

	fmt.Println("Git repository      OK")
	fmt.Printf("Staged changes      OK, %d diff bytes\n", len(data.Diff))
	printDoctorLLMs(cfg)
	fmt.Printf("Language            %s\n", cfg.Language)
	fmt.Printf("Proxy               %s\n", statusValue(cfg.Proxy))
	fmt.Printf("Max diff bytes      %d\n", cfg.MaxDiffBytes)
	fmt.Printf("Max tokens          %d\n", cfg.MaxTokens)
	fmt.Printf("Timeout seconds     %d\n", cfg.TimeoutSec)
	if path, err := config.Path(); err == nil {
		fmt.Printf("Config file         %s\n", path)
	}
	fmt.Printf("Version             %s\n", version.Version)
	fmt.Printf("Executable          %s\n", version.Executable())
	if cfg.Provider == "ollama" {
		fmt.Println("API key             not required")
	} else if cfg.APIKeyName != "" && cfg.APIKey != "" {
		if os.Getenv(cfg.APIKeyName) != "" {
			fmt.Printf("API key             OK, %s\n", cfg.APIKeyName)
		} else {
			fmt.Println("API key             OK, config file")
		}
	} else if cfg.APIKeyName != "" {
		fmt.Printf("API key             missing, %s\n", cfg.APIKeyName)
	} else {
		fmt.Println("API key             not checked")
	}
	return nil
}

func generateMessage(ctx context.Context, provider llm.Provider, cfg config.Config, data diffctx.Context) (string, error) {
	stop := ui.StartGenerating(os.Stderr)
	defer stop()

	response, err := provider.Generate(ctx, llm.GenerateRequest{Config: cfg, Diff: data})
	if err != nil {
		response, err = provider.Generate(ctx, llm.GenerateRequest{Config: cfg, Diff: data})
		if err != nil {
			return "", err
		}
	}
	return commitmsg.Render(response.Candidates[0]), nil
}

func filterOptions(cfg config.Config) diffctx.FilterOptions {
	return diffctx.FilterOptions{
		MaxBytes: cfg.MaxDiffBytes,
		Ignore:   cfg.IgnoreGlobs,
		Allow:    cfg.AllowGlobs,
	}
}

func loadConfig(ctx context.Context, client git.Client, opts Options) (config.Config, error) {
	cfg, err := config.Load(ctx, client)
	if err != nil {
		return cfg, err
	}
	if opts.Proxy == "" {
		return cfg, nil
	}
	proxy, err := config.NormalizeProxy(opts.Proxy)
	if err != nil {
		return cfg, err
	}
	cfg.Proxy = proxy
	for i := range cfg.Fallbacks {
		cfg.Fallbacks[i].Proxy = proxy
	}
	return cfg, nil
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

func printLLMEndpoints(cfg config.Config) {
	chain := cfg.Chain()
	for i, item := range chain {
		if i == 0 {
			fmt.Printf("Provider: %s\n", valueOrUnset(item.Provider))
			fmt.Printf("Model: %s\n", valueOrUnset(item.Model))
			fmt.Printf("Base URL: %s\n", valueOrUnset(item.BaseURL))
			continue
		}
		fmt.Printf("Fallback %d: %s  %s  %s\n", i+1, valueOrUnset(item.Provider), valueOrUnset(item.Model), valueOrUnset(item.BaseURL))
	}
}

func printDoctorLLMs(cfg config.Config) {
	chain := cfg.Chain()
	for i, item := range chain {
		if i == 0 {
			fmt.Printf("Provider            %s\n", statusValue(item.Provider))
			fmt.Printf("Model               %s\n", statusValue(item.Model))
			continue
		}
		fmt.Printf("Fallback %d          %s\n", i+1, item.LLMLabel())
	}
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

func confirmOverwrite(path string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s already exists. Overwrite? [y/N]\n", path)
	fmt.Print("> ")
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func addExeDirToUserPath(ctx context.Context) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	if isGoRunDir(dir) {
		return fmt.Errorf("go run uses temporary executable path; build or run git-muse.exe to update PATH")
	}
	current, err := userPath(ctx)
	if err != nil {
		return err
	}
	next := upsertGitMusePath(current, dir)
	if next == current {
		fmt.Printf("PATH already points to %s\n", dir)
	} else {
		if err := setUserPath(ctx, next); err != nil {
			return err
		}
		fmt.Printf("Updated user PATH git-muse entry to %s\n", dir)
	}
	notifyUserEnvironment()
	fmt.Println("Already-open apps keep the old PATH. Fully quit Zed/VS Code (not just the terminal) and reopen.")
	return nil
}

func userPath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "[Environment]::GetEnvironmentVariable('Path', 'User')")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func setUserPath(ctx context.Context, value string) error {
	cmd := exec.CommandContext(ctx, "reg", "add", `HKCU\Environment`, "/v", "Path", "/t", "REG_EXPAND_SZ", "/d", value, "/f")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("set user PATH: %s", msg)
	}
	return nil
}

// upsertGitMusePath puts dir first and replaces any existing git-muse install
// PATH entries (release folders like git-muse-v0.2.x-windows-amd64) instead of
// appending another version.
func upsertGitMusePath(pathValue, dir string) string {
	dir = strings.TrimSpace(dir)
	parts := strings.Split(pathValue, ";")
	out := []string{dir}
	seen := map[string]bool{strings.ToLower(filepath.Clean(dir)): true}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key := strings.ToLower(filepath.Clean(part))
		if seen[key] {
			continue
		}
		if isGitMusePathEntry(part) {
			continue
		}
		seen[key] = true
		out = append(out, part)
	}
	return strings.Join(out, ";")
}

func isGitMusePathEntry(entry string) bool {
	base := strings.ToLower(filepath.Base(filepath.Clean(strings.TrimSpace(entry))))
	return strings.Contains(base, "git-muse")
}

func isGoRunDir(dir string) bool {
	dir = strings.ToLower(filepath.Clean(dir))
	return strings.Contains(dir, string(filepath.Separator)+"go-build")
}
