package product

import (
	"go-wordpress/internal/crawler/cleanup"
	"go-wordpress/internal/crawler/description"
	"go-wordpress/internal/crawler/domain"
	"go-wordpress/internal/crawler/website"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

type Service struct {
	web    *website.Service
	logger *zap.Logger
	mu     sync.Mutex
	des    *description.Service
}

func New(logger *zap.Logger, web *website.Service) *Service {
	return &Service{
		logger: logger,
		web:    web,
		mu:     sync.Mutex{},
		des:    description.New(logger, web),
	}
}

func (s *Service) Run() error {

	if err := s.List(); err != nil {
		s.logger.Error("List failed", zap.Error(err))
		return err
	}

	web, err := s.des.FetchDescriptions(s.web)
	if err != nil {
		s.logger.Error("FetchDescriptions failed", zap.Error(err))
		return err
	}
	s.web = web

	cleanup.Cleans(s.web.Categories)
	return nil
}

func (s *Service) List() error {

	for index, cat := range s.web.Categories {
		c := colly.NewCollector(
			colly.AllowedDomains(s.web.Domain),
			colly.Async(true), // 1. Enable Native Parallelism
		)

		// 2. Configure limits to avoid being banned
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*" + s.web.Domain + "*",
			Parallelism: 4,
			Delay:       500 * time.Millisecond,
		})

		products := []domain.Product{}

		// Use configured selectors from web.ProductList
		c.OnHTML(s.web.ProductList.Item, func(e *colly.HTMLElement) {
			p := domain.Product{
				Title: e.ChildText(s.web.ProductList.Fields["title"].Selector),
				Price: e.ChildText(s.web.ProductList.Fields["price"].Selector),
				Link:  e.ChildAttr(s.web.ProductList.Fields["link"].Selector, s.web.ProductList.Fields["link"].Attr),
				Image: e.ChildAttr(s.web.ProductList.Fields["image"].Selector, s.web.ProductList.Fields["image"].Attr),
			}

			if p.Link != "" {
				// 3. Use Mutex to safely append in parallel
				s.mu.Lock()
				products = append(products, p)
				s.mu.Unlock()
			} else {
				s.logger.Debug("Product link is not valid", zap.Any("product", p))
			}
		})

		// Handle Pagination
		c.OnHTML(s.web.Pagination.Next, func(e *colly.HTMLElement) {
			nextPage := e.Attr("href")
			s.logger.Info("Visiting next page", zap.String("url", nextPage))
			e.Request.Visit(nextPage)
		})

		// 4. Handle errors for individual requests
		c.OnError(func(r *colly.Response, err error) {
			s.logger.Warn("Request failed",
				zap.String("url", r.Request.URL.String()),
				zap.Int("status", r.StatusCode),
				zap.Error(err),
			)
		})

		err := c.Visit(cat.Link)
		if err != nil {
			// Log the error and move to the NEXT category instead of returning
			s.logger.Error("Visit failed, skipping category",
				zap.String("link", cat.Link),
				zap.Error(err),
			)
			continue // This skips the current iteration and moves to the next category
		}

		//IMPORTANT: Wait for all async jobs in this category to finish
		c.Wait()

		s.web.Categories[index].Products = products
	}
	return nil
}

func (s *Service) CleanedProducts() []domain.Category { return s.web.Categories }
