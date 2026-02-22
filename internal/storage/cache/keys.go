package cache

import (
	"context"
	"fmt"
)

func (s *Store) KeyProduct(ID int32) string {
	return s.prefix + ":product:" + fmt.Sprint(ID)
}
func (s *Store) KeyAllProducts(limit, offset int64) string {
	return s.prefix + ":products:all:" + fmt.Sprint(limit) + ":" + fmt.Sprint(offset)
}

func (r *Store) DeleteAllProductCache(ctx context.Context) error {
	return r.deleteByPattern(ctx, r.prefix+":products:all:*")
}

func (s *Store) KeyWebsite(ID int32) string {
	return s.prefix + ":website:" + fmt.Sprint(ID)
}

func (s *Store) KeyAllWebsites(limit, offset int64) string {
	return s.prefix + ":websites:all:" + fmt.Sprint(limit) + ":" + fmt.Sprint(offset)
}

func (s *Store) DeleteAllWebsiteCache(ctx context.Context) {
	s.deleteByPattern(ctx, s.prefix+":websites:all:*")
}

func (s *Store) KeyCategory(ID int32) string {
	return s.prefix + ":category:" + fmt.Sprint(ID)
}

func (s *Store) KeyAllCategories(limit, offset int64) string {
	return s.prefix + ":categories:all:" + fmt.Sprint(limit) + ":" + fmt.Sprint(offset)
}

func (r *Store) DeleteAllCategoryCache(ctx context.Context) error {
	return r.deleteByPattern(ctx, r.prefix+":categories:all:*")
}

func (s *Store) KeyConfig(ID int32) string {
	return s.prefix + ":config:" + fmt.Sprint(ID)
}

func (s *Store) KeyAllConfigs(limit, offset int64) string {
	return s.prefix + ":configs:all:" + fmt.Sprint(limit) + ":" + fmt.Sprint(offset)
}

func (r *Store) DeleteAllConfigCache(ctx context.Context) error {
	return r.deleteByPattern(ctx, r.prefix+":configs:all:*")
}
