package website

import (
	"context"
	"encoding/json"
	"fmt"
	"go-wordpress/internal/crawler/domain"
	"go-wordpress/internal/storage/sql/sqlc"

	"go.uber.org/zap"
)

type Service struct {
	Websites []domain.Website
	logger   *zap.Logger
	query    *sqlc.Queries
}

func New(logger *zap.Logger, query *sqlc.Queries) (*Service, error) {
	ws := Service{
		logger: logger,
		query:  query,
	}
	err := ws.loadWebsites(context.Background())
	if err != nil {
		logger.Error("failed to fill websites from database", zap.Error(err))
		return nil, err
	}
	return &ws, nil
}

func (ws *Service) loadWebsites(ctx context.Context) error {
	websites, err := ws.query.ListActiveWebsites(ctx, sqlc.ListActiveWebsitesParams{
		Limit:  domain.DefaultLimit,
		Offset: domain.DefaultOffset,
	})
	if err != nil {
		return fmt.Errorf("failed to list active websites: %w", err)
	}

	if len(websites) == 0 {
		ws.logger.Info("no active websites found")
		return nil
	}

	configsByWebsite, err := ws.loadConfigs(ctx, websites)
	if err != nil {
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	ws.Websites = ws.buildWebsites(websites, configsByWebsite)
	return nil
}

func (ws *Service) loadConfigs(
	ctx context.Context,
	websites []sqlc.Website,
) (map[int32]*domain.CFG, error) {
	ids := extractWebsiteIDs(websites)

	configs, err := ws.query.ListActiveConfigsByWebsites(ctx, sqlc.ListActiveConfigsByWebsitesParams{
		WebsiteIds: ids,
		Limit:      domain.DefaultLimit,
		Offset:     domain.DefaultOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list active configs: %w", err)
	}

	configsByWebsite := make(map[int32]*domain.CFG, len(configs))
	for _, cfg := range configs {
		if cfg.Key != domain.ScraperConfigKey {
			continue
		}
		// Rows are ordered newest first, so the first one parsed wins.
		if _, exists := configsByWebsite[cfg.WebsiteID]; exists {
			continue
		}

		var parsed domain.CFG
		if err := json.Unmarshal(cfg.Value, &parsed); err != nil {
			ws.logger.Warn(
				"failed to unmarshal config, skipping website",
				zap.Int32("website_id", cfg.WebsiteID),
				zap.Error(err),
			)
			continue
		}
		configsByWebsite[cfg.WebsiteID] = &parsed
	}

	return configsByWebsite, nil
}

func (ws *Service) buildWebsites(
	websites []sqlc.Website,
	configsByWebsite map[int32]*domain.CFG,
) []domain.Website {
	result := make([]domain.Website, 0, len(websites))

	for _, w := range websites {
		cfg := configsByWebsite[w.ID]
		if cfg == nil {
			ws.logger.Warn(
				"no config found for website, skipping",
				zap.Int32("website_id", w.ID),
				zap.String("website_name", w.Name),
			)
			continue
		}

		result = append(result, domain.Website{
			ID:     w.ID,
			Name:   w.Name,
			Domain: w.Domain,
			CFG:    *cfg,
		})
	}

	return result
}

func extractWebsiteIDs(websites []sqlc.Website) []int32 {
	ids := make([]int32, len(websites))
	for i, w := range websites {
		ids[i] = w.ID
	}
	return ids
}

func (ws *Service) WebsiteService() []domain.Website {
	return ws.Websites
}
