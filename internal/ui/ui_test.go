package ui

import "testing"

func TestEditorProgramHandlesQuotedPath(t *testing.T) {
	got := editorProgram(`"C:\Program Files\Vim\vim.exe" --nofork`)
	want := `C:\Program Files\Vim\vim.exe`
	if got != want {
		t.Fatalf("editorProgram() = %q, want %q", got, want)
	}
}

func TestEditorPartsKeepsQuotedProgramAndArgsSeparate(t *testing.T) {
	program, args := editorParts(`"C:\Program Files\Code\bin\code.cmd" --wait --reuse-window`)
	if program != `C:\Program Files\Code\bin\code.cmd` {
		t.Fatalf("program = %q", program)
	}
	if len(args) != 2 || args[0] != "--wait" || args[1] != "--reuse-window" {
		t.Fatalf("args = %#v", args)
	}
}
