package service

import (
	"context"

	"go-wordpress/internal/category/dto"
	"go-wordpress/internal/config"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/cache"
	"go-wordpress/internal/storage/sql/sqlc"

	"go.uber.org/zap"
)

type Category struct {
	query  *sqlc.Queries
	log    *zap.Logger
	memory *cache.Store
	cfg    *config.Config
}

func New(q *sqlc.Queries, log *zap.Logger, memory *cache.Store, cfg *config.Config) *Category {
	return &Category{
		query:  q,
		log:    log,
		memory: memory,
		cfg:    cfg,
	}
}

func (s *Category) Create(ctx context.Context, req sqlc.CreateCategoryParams) (*sqlc.Category, error) {
	category, err := s.query.CreateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyCategory(category.ID), category, s.cfg.Redis.DefaultTTL)
	if err := s.memory.DeleteAllCategoryCache(ctx); err != nil {
		s.log.Warn("failed to invalidate category list cache", zap.Error(err))
	}
	return &category, nil
}

func (s *Category) Update(ctx context.Context, req sqlc.UpdateCategoryParams) (*sqlc.Category, error) {
	category, err := s.query.UpdateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyCategory(category.ID), category, s.cfg.Redis.DefaultTTL)
	if err := s.memory.DeleteAllCategoryCache(ctx); err != nil {
		s.log.Warn("failed to invalidate category list cache", zap.Error(err))
	}
	return &category, nil
}

func (s *Category) Delete(ctx context.Context, id int32) error {
	s.memory.Delete(ctx, s.memory.KeyCategory(id))
	if err := s.memory.DeleteAllCategoryCache(ctx); err != nil {
		s.log.Warn("failed to invalidate category list cache", zap.Error(err))
	}
	return s.query.DeleteCategory(ctx, id)
}

func (s *Category) GetCategoryByID(ctx context.Context, id int32) (*sqlc.Category, error) {
	var category sqlc.Category
	err := s.memory.Get(ctx, s.memory.KeyCategory(id), &category)
	if err != nil {
		category, err = s.query.GetCategoryByID(ctx, id)
		if err != nil {
			return nil, err
		}
		s.memory.Set(ctx, s.memory.KeyCategory(category.ID), category, s.cfg.Redis.DefaultTTL)
	}
	return &category, nil
}

func (s *Category) ListCategories(ctx context.Context, pagination shared.Pagination) (dto.CategoriesResponse, error) {
	var resp dto.CategoriesResponse
	if err := s.memory.Get(ctx, s.memory.KeyAllCategories(pagination.Limit, pagination.Offset), &resp); err == nil {
		return resp, nil
	}
	if pagination.Limit == 0 {
		pagination.Limit = 10
	}
	categories, err := s.query.ListAllCategories(ctx, sqlc.ListAllCategoriesParams{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	resp = make(dto.CategoriesResponse, 0, len(categories))
	for _, c := range categories {
		resp = append(resp, c)
	}
	s.memory.Set(ctx, s.memory.KeyAllCategories(pagination.Limit, pagination.Offset), resp, s.cfg.Redis.DefaultTTL)
	return resp, nil
}
