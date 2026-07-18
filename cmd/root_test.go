package cmd

import "testing"

func TestParseGlobalFlags(t *testing.T) {
	flags, err := parseGlobalFlags([]string{"--help"})
	if err != nil || !flags.Help || flags.Version || len(flags.Rest) != 0 {
		t.Fatalf("help parse: %+v err=%v", flags, err)
	}
	flags, err = parseGlobalFlags([]string{"--version"})
	if err != nil || flags.Help || !flags.Version || len(flags.Rest) != 0 {
		t.Fatalf("version parse: %+v err=%v", flags, err)
	}
	flags, err = parseGlobalFlags([]string{"generate", "-V"})
	if err != nil || flags.Help || !flags.Version || len(flags.Rest) != 1 || flags.Rest[0] != "generate" {
		t.Fatalf("mixed parse: %+v err=%v", flags, err)
	}
}

func TestParseGlobalFlagsProxy(t *testing.T) {
	flags, err := parseGlobalFlags([]string{"--proxy", "127.0.0.1:7890", "generate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.Proxy != "127.0.0.1:7890" {
		t.Fatalf("proxy = %q", flags.Proxy)
	}
	if len(flags.Rest) != 1 || flags.Rest[0] != "generate" {
		t.Fatalf("rest = %#v", flags.Rest)
	}

	flags, err = parseGlobalFlags([]string{"--proxy=http://127.0.0.1:7890"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.Proxy != "http://127.0.0.1:7890" {
		t.Fatalf("proxy = %q", flags.Proxy)
	}

	if _, err := parseGlobalFlags([]string{"--proxy"}); err == nil {
		t.Fatalf("expected missing proxy value error")
	}
}
