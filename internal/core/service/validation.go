package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/foden/cdc/internal/core/ports"
	cdcerrors "github.com/foden/cdc/pkg/errors"
)

func validateUniqueSource(ctx context.Context, store ports.Store, candidate *ports.SourceConfig) error {
	existingSources, err := store.ListSources(ctx)
	if err != nil {
		return err
	}
	candidateID := normalizeToken(candidate.InstanceID)
	candidateName := normalizeToken(candidate.Name)
	candidateEndpoint := sourceEndpointKey(candidate)
	for _, existing := range existingSources {
		if existing == nil {
			continue
		}
		if candidateID != "" && normalizeToken(existing.InstanceID) == candidateID {
			return fmt.Errorf("%w: source instance_id %q already exists", cdcerrors.ErrDuplicateConfig, candidate.InstanceID)
		}
		if candidateName != "" && normalizeToken(existing.Name) == candidateName {
			return fmt.Errorf("%w: source name %q already exists", cdcerrors.ErrDuplicateConfig, candidate.Name)
		}
		if candidateEndpoint != "" && sourceEndpointKey(existing) == candidateEndpoint {
			return fmt.Errorf("%w: source endpoint already exists", cdcerrors.ErrDuplicateConfig)
		}
	}
	return nil
}

func validateUniqueSink(ctx context.Context, store ports.Store, candidate *ports.SinkConfig) error {
	existingSinks, err := store.ListSinks(ctx)
	if err != nil {
		return err
	}
	candidateID := normalizeToken(candidate.InstanceID)
	candidateName := normalizeToken(candidate.Name)
	candidateEndpoint := sinkEndpointKey(candidate)
	for _, existing := range existingSinks {
		if existing == nil {
			continue
		}
		if candidateID != "" && normalizeToken(existing.InstanceID) == candidateID {
			return fmt.Errorf("%w: sink instance_id %q already exists", cdcerrors.ErrDuplicateConfig, candidate.InstanceID)
		}
		if candidateName != "" && normalizeToken(existing.Name) == candidateName {
			return fmt.Errorf("%w: sink name %q already exists", cdcerrors.ErrDuplicateConfig, candidate.Name)
		}
		if candidateEndpoint != "" && sinkEndpointKey(existing) == candidateEndpoint {
			return fmt.Errorf("%w: sink endpoint already exists", cdcerrors.ErrDuplicateConfig)
		}
	}
	return nil
}

func sourceEndpointKey(cfg *ports.SourceConfig) string {
	if cfg == nil {
		return ""
	}
	return strings.Join([]string{
		normalizeToken(cfg.Type),
		normalizeToken(cfg.Host),
		fmt.Sprintf("%d", cfg.Port),
		normalizeToken(cfg.Database),
		normalizeToken(cfg.Username),
	}, "|")
}

func sinkEndpointKey(cfg *ports.SinkConfig) string {
	if cfg == nil {
		return ""
	}
	if normalizeToken(cfg.Type) == "elasticsearch" {
		urls := make([]string, 0, len(cfg.URL))
		for _, url := range cfg.URL {
			if normalized := normalizeToken(url); normalized != "" {
				urls = append(urls, normalized)
			}
		}
		sort.Strings(urls)
		return strings.Join([]string{
			"elasticsearch",
			strings.Join(urls, ","),
			normalizeToken(cfg.IndexPrefix),
			hasSecret(cfg.APIKey),
		}, "|")
	}
	return strings.Join([]string{
		normalizeToken(cfg.Type),
		normalizeToken(cfg.Host),
		fmt.Sprintf("%d", cfg.Port),
		normalizeToken(cfg.Database),
		normalizeToken(cfg.Username),
	}, "|")
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func hasSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return "no-secret"
	}
	return "has-secret"
}
