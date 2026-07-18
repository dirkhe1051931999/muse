package llm

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func newHTTPClient(timeoutSec int, proxy string) (*http.Client, error) {
	if timeoutSec <= 0 {
		timeoutSec = 180
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	if proxy == "" {
		return client, nil
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy %q: %w", proxy, err)
	}
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()
	transport.Proxy = http.ProxyURL(proxyURL)
	client.Transport = transport
	return client, nil
}
