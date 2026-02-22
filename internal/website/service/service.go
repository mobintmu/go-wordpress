package service

import (
	"context"

	"go-wordpress/internal/config"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/cache"
	"go-wordpress/internal/storage/sql/sqlc"
	"go-wordpress/internal/website/dto"

	"go.uber.org/zap"
)

type Website struct {
	query  *sqlc.Queries
	log    *zap.Logger
	memory *cache.Store
	cfg    *config.Config
}

func New(q *sqlc.Queries, log *zap.Logger, memory *cache.Store, cfg *config.Config) *Website {
	return &Website{
		query:  q,
		log:    log,
		memory: memory,
		cfg:    cfg,
	}
}

func (s *Website) Create(ctx context.Context, req sqlc.CreateWebsiteParams) (*sqlc.CreateWebsiteRow, error) {
	website, err := s.query.CreateWebsite(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyWebsite(website.ID), website, s.cfg.Redis.DefaultTTL)
	s.memory.DeleteAllWebsiteCache(ctx)
	return &website, nil
}

func (s *Website) Update(ctx context.Context, req sqlc.UpdateWebsiteParams) (*sqlc.Website, error) {
	website, err := s.query.UpdateWebsite(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyWebsite(website.ID), website, s.cfg.Redis.DefaultTTL)
	s.memory.DeleteAllWebsiteCache(ctx)
	return &website, nil
}

func (s *Website) Delete(ctx context.Context, id int32) error {
	s.memory.Delete(ctx, s.memory.KeyWebsite(id))
	s.memory.DeleteAllWebsiteCache(ctx)
	return s.query.DeleteWebsite(ctx, id)
}

func (s *Website) GetWebsiteByID(ctx context.Context, id int32) (*sqlc.Website, error) {
	var website sqlc.Website
	err := s.memory.Get(ctx, s.memory.KeyWebsite(id), &website)
	if err != nil {
		website, err = s.query.GetWebsiteByID(ctx, id)
		if err != nil {
			return nil, err
		}
		s.memory.Set(ctx, s.memory.KeyWebsite(website.ID), website, s.cfg.Redis.DefaultTTL)
	}
	return &website, nil
}

func (s *Website) ListWebsites(ctx context.Context, pagination shared.Pagination) (dto.WebsitesResponse, error) {
	var resp dto.WebsitesResponse
	if err := s.memory.Get(ctx, s.memory.KeyAllWebsites(pagination.Limit, pagination.Offset), &resp); err == nil {
		return resp, nil
	}
	if pagination.Limit == 0 {
		pagination.Limit = 10
	}
	websites, err := s.query.ListAllWebsites(ctx, sqlc.ListAllWebsitesParams{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	resp = make(dto.WebsitesResponse, 0, len(websites))
	for _, w := range websites {
		resp = append(resp, w)
	}
	s.memory.Set(ctx, s.memory.KeyAllWebsites(pagination.Limit, pagination.Offset), resp, s.cfg.Redis.DefaultTTL)
	return resp, nil
}
