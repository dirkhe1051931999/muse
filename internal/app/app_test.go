package app

import "testing"

func TestUpsertGitMusePathReplacesOlderReleaseDirs(t *testing.T) {
	current := `D:\oolaf-ambisonic\git-muse-v0.2.8-windows-amd64;C:\Windows\System32;D:\oolaf-ambisonic\git-muse-v0.2.7-windows-amd64`
	got := upsertGitMusePath(current, `D:\oolaf-ambisonic\git-muse-v0.2.9-windows-amd64`)
	want := `D:\oolaf-ambisonic\git-muse-v0.2.9-windows-amd64;C:\Windows\System32`
	if got != want {
		t.Fatalf("upsertGitMusePath() = %q, want %q", got, want)
	}
}

func TestUpsertGitMusePathAddsWhenMissing(t *testing.T) {
	got := upsertGitMusePath(`C:\Windows\System32;tools`, `D:\tools\git-muse`)
	want := `D:\tools\git-muse;C:\Windows\System32;tools`
	if got != want {
		t.Fatalf("upsertGitMusePath() = %q, want %q", got, want)
	}
}

func TestUpsertGitMusePathKeepsUnrelatedDirs(t *testing.T) {
	got := upsertGitMusePath(`C:\Users\Administrator\bin;tools`, `C:\Users\Administrator\bin`)
	want := `C:\Users\Administrator\bin;tools`
	if got != want {
		t.Fatalf("upsertGitMusePath() = %q, want %q", got, want)
	}
}

func TestIsGitMusePathEntry(t *testing.T) {
	if !isGitMusePathEntry(`D:\oolaf-ambisonic\git-muse-v0.2.8-windows-amd64`) {
		t.Fatalf("expected release dir to match")
	}
	if isGitMusePathEntry(`C:\Users\Administrator\AppData\Local\codegraph\current\bin`) {
		t.Fatalf("unrelated bin should not match")
	}
}
