package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"git-muse/internal/config"
)

type Failover struct {
	Providers        []Provider
	Labels           []string
	RateLimitRetries int
	Log              io.Writer
}

func (p Failover) Name() string {
	if len(p.Providers) == 0 {
		return "failover"
	}
	return p.Providers[0].Name()
}

func (p Failover) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if len(p.Providers) == 0 {
		return GenerateResponse{}, fmt.Errorf("no LLM providers configured")
	}
	log := p.Log
	if log == nil {
		log = os.Stderr
	}
	retries := config.NormalizeRateLimitRetries(p.RateLimitRetries)
	var last error
	for i, provider := range p.Providers {
		label := labelAt(p.Labels, i, provider.Name())
		for attempt := 1; attempt <= retries; attempt++ {
			resp, err := provider.Generate(ctx, req)
			if err == nil {
				return resp, nil
			}
			last = err
			if !IsRateLimited(err) {
				return GenerateResponse{}, err
			}
			if attempt < retries {
				fmt.Fprintf(log, "\ngit-muse: LLM rate limited: %s (%s); retry %d/%d\n", label, shortErr(err), attempt, retries)
				continue
			}
			if i+1 >= len(p.Providers) {
				return GenerateResponse{}, err
			}
			to := labelAt(p.Labels, i+1, p.Providers[i+1].Name())
			fmt.Fprintf(log, "\ngit-muse: LLM rate limited: %s (%s); switching to %s after %d attempts\n", label, shortErr(err), to, retries)
		}
	}
	return GenerateResponse{}, last
}

func newChain(cfgs []config.Config, rateLimitRetries int) (Provider, error) {
	if len(cfgs) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}
	retries := config.NormalizeRateLimitRetries(rateLimitRetries)
	providers := make([]Provider, 0, len(cfgs))
	labels := make([]string, 0, len(cfgs))
	for i, cfg := range cfgs {
		p, err := newOne(cfg)
		if err != nil {
			if i == 0 {
				return nil, err
			}
			fmt.Fprintf(os.Stderr, "git-muse: skip fallback %s: %v\n", cfg.LLMLabel(), err)
			continue
		}
		providers = append(providers, p)
		labels = append(labels, cfg.LLMLabel())
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}
	if len(providers) == 1 && retries == 1 {
		return providers[0], nil
	}
	return Failover{Providers: providers, Labels: labels, RateLimitRetries: retries}, nil
}

func labelAt(labels []string, i int, fallback string) string {
	if i >= 0 && i < len(labels) && labels[i] != "" {
		return labels[i]
	}
	return fallback
}

func shortErr(err error) string {
	var rl *RateLimitError
	if errors.As(err, &rl) && rl.Status != "" {
		return rl.Status
	}
	msg := err.Error()
	if len(msg) > 120 {
		return msg[:120] + "..."
	}
	return msg
}
