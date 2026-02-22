package service

import (
	"context"

	app "go-wordpress/internal/config"
	"go-wordpress/internal/configs/dto"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/cache"
	"go-wordpress/internal/storage/sql/sqlc"

	"go.uber.org/zap"
)

type Config struct {
	query  *sqlc.Queries
	log    *zap.Logger
	memory *cache.Store
	cfg    *app.Config
}

func New(q *sqlc.Queries, log *zap.Logger, memory *cache.Store, cfg *app.Config) *Config {
	return &Config{
		query:  q,
		log:    log,
		memory: memory,
		cfg:    cfg,
	}
}

const (
	MessageFailedToInvalidateConfigListCache = "failed to invalidate config list cache"
)

func (s *Config) Create(ctx context.Context, req sqlc.CreateConfigParams) (*sqlc.CreateConfigRow, error) {
	cfg, err := s.query.CreateConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyConfig(cfg.ID), cfg, s.cfg.Redis.DefaultTTL)
	if err := s.memory.DeleteAllConfigCache(ctx); err != nil {
		s.log.Warn(MessageFailedToInvalidateConfigListCache, zap.Error(err))
	}
	return &cfg, nil
}

func (s *Config) Update(ctx context.Context, req sqlc.UpdateConfigParams) (*sqlc.UpdateConfigRow, error) {
	cfg, err := s.query.UpdateConfig(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyConfig(cfg.ID), cfg, s.cfg.Redis.DefaultTTL)
	if err := s.memory.DeleteAllConfigCache(ctx); err != nil {
		s.log.Warn(MessageFailedToInvalidateConfigListCache, zap.Error(err))
	}
	return &cfg, nil
}

func (s *Config) Delete(ctx context.Context, id int32) error {
	s.memory.Delete(ctx, s.memory.KeyConfig(id))
	if err := s.memory.DeleteAllConfigCache(ctx); err != nil {
		s.log.Warn(MessageFailedToInvalidateConfigListCache, zap.Error(err))
	}
	return s.query.DeleteConfig(ctx, id)
}

func (s *Config) GetConfigByID(ctx context.Context, id int32) (*sqlc.GetConfigByIDRow, error) {
	var cfg sqlc.GetConfigByIDRow
	err := s.memory.Get(ctx, s.memory.KeyConfig(id), &cfg)
	if err != nil {
		cfg, err = s.query.GetConfigByID(ctx, id)
		if err != nil {
			return nil, err
		}
		s.memory.Set(ctx, s.memory.KeyConfig(cfg.ID), cfg, s.cfg.Redis.DefaultTTL)
	}
	return &cfg, nil
}

func (s *Config) ListConfigs(ctx context.Context, pagination shared.Pagination) (dto.ConfigsResponse, error) {
	var resp dto.ConfigsResponse
	if err := s.memory.Get(ctx, s.memory.KeyAllConfigs(pagination.Limit, pagination.Offset), &resp); err == nil {
		return resp, nil
	}
	if pagination.Limit == 0 {
		pagination.Limit = 10
	}
	configs, err := s.query.ListAllConfigs(ctx, sqlc.ListAllConfigsParams{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	resp = make(dto.ConfigsResponse, 0, len(configs))
	for _, c := range configs {
		resp = append(resp, c)
	}
	s.memory.Set(ctx, s.memory.KeyAllConfigs(pagination.Limit, pagination.Offset), resp, s.cfg.Redis.DefaultTTL)
	return resp, nil
}
