package llm

import (
	"net/http"
	"testing"
)

func TestNewHTTPClientWithProxy(t *testing.T) {
	client, err := newHTTPClient(30, "http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy == nil {
		t.Fatalf("expected transport with proxy")
	}
	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	u, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("proxy lookup: %v", err)
	}
	if u.String() != "http://127.0.0.1:7890" {
		t.Fatalf("proxy URL = %q", u.String())
	}
}

func TestNewHTTPClientWithoutProxy(t *testing.T) {
	client, err := newHTTPClient(30, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Transport != nil {
		t.Fatalf("expected default transport")
	}
}
