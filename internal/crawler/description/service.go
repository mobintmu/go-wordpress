package description

import (
	"go-wordpress/internal/crawler/website"
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

func New(logger *zap.Logger, web *website.Service) *Service {
	return &Service{
		logger:     logger,
		mu:         sync.Mutex{},
		batchSize:  4,
		sleepBatch: 2 * time.Second,
	}
}

func (s *Service) FetchDescriptions(web *website.Service) (w *website.Service, err error) {
	for categoryIndex, cat := range web.Categories {
		// Get valid products
		var validProducts []int
		for i := range cat.Products {
			if web.Categories[categoryIndex].Products[i].Link != "" {
				validProducts = append(validProducts, i)
			}
		}

		s.logger.Info("Category processing started",
			zap.String("category", web.Categories[categoryIndex].Name),
			zap.Int("total_products", len(validProducts)),
		)

		// Process products in batches
		err := s.processBatches(web, categoryIndex, validProducts)
		if err != nil {
			s.logger.Error("Failed to process batches",
				zap.String("category", web.Categories[categoryIndex].Name),
				zap.Error(err),
			)
		}
	}
	return web, nil
}

func (s *Service) processBatches(web *website.Service, categoryIndex int, validProducts []int) error {
	for batchStart := 0; batchStart < len(validProducts); batchStart += s.batchSize {
		batchEnd := batchStart + s.batchSize
		if batchEnd > len(validProducts) {
			batchEnd = len(validProducts)
		}

		s.logger.Info("Processing batch",
			zap.String("category", web.Categories[categoryIndex].Name),
			zap.Int("batch_start", batchStart),
			zap.Int("batch_end", batchEnd),
			zap.Int("total_products", len(validProducts)),
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
			web.ProductDetail.Description.Selector,
			func(e *colly.HTMLElement) {
				desc := strings.TrimSpace(e.Text)

				// Get context data
				idxAny := e.Request.Ctx.Get("index")
				catIdxAny := e.Request.Ctx.Get("categoryIndex")

				idx, _ := strconv.Atoi(idxAny)
				catIdx, _ := strconv.Atoi(catIdxAny)

				if desc == "" {
					s.logger.Warn("empty description", zap.String("url", e.Request.URL.String()))
					return
				}

				s.mu.Lock()
				web.Categories[catIdx].Products[idx].Description = desc
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

		for _, productIdx := range validProducts[batchStart:batchEnd] {
			wg.Add(1)
			go func(productIndex int, categoryIdx int) {
				defer wg.Done()

				semaphore <- struct{}{}        // Acquire slot
				defer func() { <-semaphore }() // Release slot

				ctx := colly.NewContext()
				ctx.Put("index", strconv.Itoa(productIndex))
				ctx.Put("categoryIndex", strconv.Itoa(categoryIdx))

				if err := c.Request(
					"GET",
					web.Categories[categoryIdx].Products[productIndex].Link,
					nil,
					ctx,
					nil,
				); err != nil {
					s.logger.Warn(
						"Failed to visit product",
						zap.String("url", web.Categories[categoryIdx].Products[productIndex].Link),
						zap.Error(err),
					)
				} else {
					s.logger.Info("Visiting product page",
						zap.String("url", web.Categories[categoryIdx].Products[productIndex].Link),
					)
				}
			}(productIdx, categoryIndex)
		}

		wg.Wait()
		c.Wait() // Wait for all async jobs to complete

		s.logger.Info("Batch completed",
			zap.String("category", web.Categories[categoryIndex].Name),
			zap.Int("batch_start", batchStart),
			zap.Int("batch_end", batchEnd),
		)

		// Add delay between batches to be nice to the server
		time.Sleep(s.sleepBatch)
	}

	return nil
}
