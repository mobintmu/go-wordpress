package service

import (
	"context"
	"go-wordpress/internal/config"
	"go-wordpress/internal/product/dto"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/cache"
	"go-wordpress/internal/storage/sql/sqlc"

	"go.uber.org/zap"
)

type Product struct {
	query  *sqlc.Queries
	log    *zap.Logger
	memory *cache.Store
	cfg    *config.Config
}

func New(q *sqlc.Queries,
	log *zap.Logger,
	memory *cache.Store,
	cfg *config.Config) *Product {
	return &Product{
		query:  q,
		log:    log,
		memory: memory,
		cfg:    cfg,
	}
}

func (s *Product) Create(ctx context.Context, req sqlc.CreateProductParams) (*sqlc.Product, error) {
	product, err := s.query.CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyProduct(product.ID), product, s.cfg.Redis.DefaultTTL)
	s.memory.DeleteAllProductCache(ctx)
	return &product, nil
}

func (s *Product) Update(ctx context.Context, req sqlc.UpdateProductParams) (*sqlc.Product, error) {
	product, err := s.query.UpdateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	s.memory.Set(ctx, s.memory.KeyProduct(product.ID), product, s.cfg.Redis.DefaultTTL)
	s.memory.DeleteAllProductCache(ctx)
	return &product, nil
}

func (s *Product) Delete(ctx context.Context, id int32) error {
	s.memory.Delete(ctx, s.memory.KeyProduct(id))
	s.memory.DeleteAllProductCache(ctx)
	return s.query.DeleteProduct(ctx, id)
}

func (s *Product) GetProductByID(ctx context.Context, id int32) (*sqlc.Product, error) {
	var product sqlc.Product
	err := s.memory.Get(ctx, s.memory.KeyProduct(id), &product)
	if err != nil {
		product, err = s.query.GetProductByID(ctx, id)
		if err != nil {
			return nil, err
		}
		s.memory.Set(ctx, s.memory.KeyProduct(product.ID), product, s.cfg.Redis.DefaultTTL)
	}
	return &product, nil
}

func (s *Product) ListProducts(ctx context.Context, pagination shared.Pagination) (dto.ProductsResponse, error) {
	var resp dto.ProductsResponse
	if err := s.memory.Get(ctx, s.memory.KeyAllProducts(pagination.Limit, pagination.Offset), &resp); err == nil {
		return resp, nil
	}
	if pagination.Limit == 0 {
		pagination.Limit = 10
	}
	products, err := s.query.ListAllProducts(ctx, sqlc.ListAllProductsParams{
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	resp = make([]sqlc.Product, 0, len(products))
	for _, product := range products {
		resp = append(resp, product)
	}
	s.memory.Set(ctx, s.memory.KeyAllProducts(pagination.Limit, pagination.Offset), resp, s.cfg.Redis.DefaultTTL)
	return resp, nil
}
