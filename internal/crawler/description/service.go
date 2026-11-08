package description

import (
	"fmt"
	"go-wordpress/internal/crawler/domain"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

type Service struct {
	logger     *zap.Logger
	mu         sync.Mutex
	batchSize  int
	sleepBatch time.Duration
}

func New(logger *zap.Logger) *Service {
	return &Service{
		logger:     logger,
		mu:         sync.Mutex{},
		batchSize:  4,
		sleepBatch: 2 * time.Second,
	}
}

func (s *Service) validateWeb(web *domain.Website) error {
	if web == nil {
		return fmt.Errorf("website cannot be nil")
	}
	if web.Domain == "" {
		return fmt.Errorf("website domain cannot be empty")
	}
	if len(web.Categories) == 0 {
		return fmt.Errorf("website must have at least 1 category for description fetching")
	}
	if web.CFG.StartURL == "" {
		return fmt.Errorf("website start URL cannot be empty")
	}
	if web.CFG.ProductDetail == nil || web.CFG.ProductDetail.Description.Selector == "" {
		return fmt.Errorf("product description selector cannot be empty")
	}
	return nil
}

func (s *Service) removeInvalidProducts(web *domain.Website) {
	for i, cat := range web.Categories {
		var validProducts []domain.Product
		for _, prod := range cat.Products {
			if prod.Link != "" {
				validProducts = append(validProducts, prod)
			}
		}
		web.Categories[i].Products = validProducts
	}
}

// FetchDescriptions fetches product descriptions for all products in every category of the website. It processes products in batches to control concurrency and avoid overwhelming the target server.
func (s *Service) FetchDescriptions(web *domain.Website) error {
	err := s.validateWeb(web)
	if err != nil {
		s.logger.Error("website validation failed", zap.Error(err))
		return err
	}
	s.removeInvalidProducts(web)

	// Process products in batches
	for catIndex := range web.Categories {
		if err := s.processBatches(web, catIndex); err != nil {
			s.logger.Error("Failed to process batches",
				zap.String("category", web.Categories[catIndex].Name),
				zap.Error(err),
			)
			return fmt.Errorf("process batches for category %q: %w", web.Categories[catIndex].Name, err)
		}
	}
	return nil
}

func (s *Service) processBatches(web *domain.Website, catIndex int) error {
	err := s.validateWeb(web)
	if err != nil {
		s.logger.Error("website validation failed", zap.Error(err))
		return err
	}
	category := &web.Categories[catIndex]
	products := category.Products
	for batchStart := 0; batchStart < len(products); batchStart += s.batchSize {
		batchEnd := batchStart + s.batchSize
		if batchEnd > len(products) {
			batchEnd = len(products)
		}

		s.logger.Info("Processing batch",
			zap.String("category", category.Name),
			zap.Int("batch_start", batchStart),
			zap.Int("batch_end", batchEnd),
			zap.Int("total_products", len(products)),
		)

		c := colly.NewCollector(
			colly.AllowedDomains(web.Domain),
			colly.Async(true),
		)

		// Control parallelism
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*" + web.Domain + "*",
			Parallelism: 4, // number of concurrent requests
			Delay:       800 * time.Millisecond,
		})

		c.SetRequestTimeout(30 * time.Second)

		// Extract description from product page
		c.OnHTML(
			web.CFG.ProductDetail.Description.Selector,
			func(e *colly.HTMLElement) {
				desc := strings.TrimSpace(e.Text)

				// Get context data
				idxAny := e.Request.Ctx.Get("index")
				idx, _ := strconv.Atoi(idxAny)

				if desc == "" {
					s.logger.Warn("empty description", zap.String("url", e.Request.URL.String()))
					return
				}

				s.mu.Lock()
				category.Products[idx].Description = desc
				s.mu.Unlock()
			},
		)

		c.OnError(func(res *colly.Response, err error) {
			s.logger.Warn("Request failed",
				zap.String("url", res.Request.URL.String()),
				zap.Error(err),
			)
		})

		// Queue batch of 4 products
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 4) // Max 4 concurrent requests

		for productIndex := batchStart; productIndex < batchEnd; productIndex++ {
			wg.Add(1)
			go func(productIndex int) {
				defer wg.Done()

				semaphore <- struct{}{}        // Acquire slot
				defer func() { <-semaphore }() // Release slot

				ctx := colly.NewContext()
				ctx.Put("index", strconv.Itoa(productIndex))

				if err := c.Request(
					"GET",
					category.Products[productIndex].Link,
					nil,
					ctx,
					nil,
				); err != nil {
					s.logger.Warn(
						"Failed to visit product",
						zap.String("url", category.Products[productIndex].Link),
						zap.Error(err),
					)
				} else {
					s.logger.Info("Visiting product page",
						zap.String("url", category.Products[productIndex].Link),
					)
				}
			}(productIndex)
		}

		wg.Wait()
		c.Wait() // Wait for all async jobs to complete

		s.logger.Info("Batch completed",
			zap.String("category", category.Name),
			zap.Int("batch_start", batchStart),
			zap.Int("batch_end", batchEnd),
		)

		// Add delay between batches to be nice to the server
		time.Sleep(s.sleepBatch)
	}

	return nil
}
