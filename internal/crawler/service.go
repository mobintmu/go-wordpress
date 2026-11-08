package crawler

import (
	"go-wordpress/internal/storage/sql/sqlc"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	query  *sqlc.Queries
	logger *zap.Logger

	// Configuration for scraping behavior
	batchSize  int
	sleepBatch time.Duration
}

// New creates a new integrated Crawler Service.
func New(logger *zap.Logger, query *sqlc.Queries) *Service {
	return &Service{
		logger:     logger,
		query:      query,
		batchSize:  4,
		sleepBatch: 2 * time.Second,
	}
}
