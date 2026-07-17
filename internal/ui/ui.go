package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"git-muse/internal/git"
)

func Choose(ctx context.Context, client git.Client, message string, regenerate func() (string, error)) (string, bool, bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Suggested commit message:")
		fmt.Println()
		fmt.Print(message)
		fmt.Println()
		fmt.Println("[c] commit  [p] commit and push  [e] edit  [r] regenerate  [q] quit")
		fmt.Print("> ")

		choice, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(choice)) {
		case "c", "commit":
			return message, true, false, nil
		case "p", "push", "commit and push":
			return message, true, true, nil
		case "e", "edit":
			edited, err := Edit(ctx, client, message)
			if err != nil {
				return "", false, false, err
			}
			if strings.TrimSpace(edited) == "" {
				return "", false, false, fmt.Errorf("commit message is empty")
			}
			return edited, true, false, nil
		case "r", "regenerate":
			next, err := regenerate()
			if err != nil {
				return "", false, false, err
			}
			message = next
		case "q", "quit", "":
			return "", false, false, nil
		default:
			fmt.Println("Unknown choice.")
		}
	}
}

func Confirm(message string, push bool) bool {
	reader := bufio.NewReader(os.Stdin)
	if push {
		fmt.Println("Commit and push with this message?")
	} else {
		fmt.Println("Commit with this message?")
	}
	fmt.Println()
	fmt.Print(message)
	fmt.Println("[y/N]")
	fmt.Print("> ")
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func Edit(ctx context.Context, client git.Client, message string) (string, error) {
	file, err := os.CreateTemp("", "git-muse-edit-*.txt")
	if err != nil {
		return "", err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.WriteString(message); err != nil {
		file.Close()
		return "", err
	}
	file.Close()

	editor := usableEditor(ctx, client)
	cmd := editorCommand(ctx, editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func usableEditor(ctx context.Context, client git.Client) string {
	editor := client.Editor(ctx)
	if editor == "" {
		return defaultEditor()
	}
	if runtime.GOOS == "windows" {
		name := editorProgram(editor)
		if _, err := exec.LookPath(name); err != nil {
			return defaultEditor()
		}
	}
	return editor
}

func defaultEditor() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

func editorProgram(editor string) string {
	program, _ := editorParts(editor)
	return program
}

func editorCommand(ctx context.Context, editor string, path string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		program, args := editorParts(editor)
		args = append(args, path)
		return exec.CommandContext(ctx, program, args...)
	}
	return exec.CommandContext(ctx, "sh", "-c", editor+" \"$1\"", "git-muse-editor", path)
}

func editorParts(editor string) (string, []string) {
	editor = strings.TrimSpace(editor)
	if strings.HasPrefix(editor, "\"") {
		end := strings.Index(editor[1:], "\"")
		if end >= 0 {
			program := editor[1 : end+1]
			return program, strings.Fields(strings.TrimSpace(editor[end+2:]))
		}
	}
	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}
