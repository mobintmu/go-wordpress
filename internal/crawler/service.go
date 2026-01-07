package crawler

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

type Service struct {
	products        []Product
	cleanedProducts []Product
	logger          *zap.Logger
	mu              sync.Mutex
}

func New(logger *zap.Logger) *Service {
	return &Service{
		logger:   logger,
		products: []Product{},
		mu:       sync.Mutex{},
	}
}

func (s *Service) Run() error {
	if err := s.ListProducts(); err != nil {
		s.logger.Error("ListProducts failed", zap.Error(err))
		return err
	}

	if err := s.FetchDescriptions(); err != nil {
		s.logger.Error("FetchDescriptions failed", zap.Error(err))
		return err
	}

	s.cleanUP()
	return nil
}

func (s *Service) ListProducts() error {
	c := colly.NewCollector(
		colly.AllowedDomains("badomjip.com"),
	)

	products := []Product{}

	c.OnHTML("div.product-grid-item", func(e *colly.HTMLElement) {
		p := Product{
			Title: e.ChildText("h3.wd-entities-title a"),
			Price: e.ChildText("span.price"),
			Link:  e.ChildAttr("a.product-image-link", "href"),
			Image: e.ChildAttr("a.product-image-link img", "src"),
		}
		// Check if the product link is valid
		if p.Link != "" {
			products = append(products, p)
		} else {
			s.logger.Info("Product link is not valid", zap.Any("product", p))
		}
	})

	err := c.Visit("https://badomjip.com/product-category/%d9%be%da%a9%db%8c%d8%ac%d9%87%d8%a7/")
	if err != nil {
		return err
	}

	s.products = products
	return nil
}

func (s *Service) cleanUP() {
	cleaned := make([]Product, 0, len(s.products))

	// regex to strip HTML tags
	reTags := regexp.MustCompile(`<[^>]*>`)

	for _, p := range s.products {
		// --- Clean Title ---
		title := reTags.ReplaceAllString(p.Title, "") // remove tags
		title = strings.TrimSpace(title)

		// --- Clean Price ---
		// Example: "499,000 تومان"
		priceStr := p.Price
		// remove currency word
		priceStr = strings.ReplaceAll(priceStr, "تومان", "")
		// remove commas and non-digits
		priceStr = strings.ReplaceAll(priceStr, ",", "")
		priceStr = strings.TrimSpace(priceStr)

		// keep only digits
		reDigits := regexp.MustCompile(`\D`)
		priceStr = reDigits.ReplaceAllString(priceStr, "")

		var priceInt int
		if priceStr != "" {
			if val, err := strconv.Atoi(priceStr); err == nil {
				priceInt = val
			}
		}

		// --- Build cleaned product ---
		cleaned = append(cleaned, Product{
			Title:       title,
			Price:       strconv.Itoa(priceInt), // or change Product.Price to int if you prefer
			Link:        p.Link,
			Image:       p.Image,
			Description: p.Description,
		})
	}
	s.cleanedProducts = cleaned
}

func (s *Service) CleanedProducts() []Product { return s.cleanedProducts }

func (s *Service) FetchDescriptions() error {
	c := colly.NewCollector(
		colly.AllowedDomains("badomjip.com"),
		colly.Async(true), // ENABLE ASYNC
	)

	// Control parallelism
	c.Limit(&colly.LimitRule{
		DomainGlob:  "badomjip.com",
		Parallelism: 4, // number of concurrent requests
		Delay:       800 * time.Millisecond,
	})

	// Extract description from product page
	c.OnHTML(
		"div.woocommerce-product-details__short-description, div.woocommerce-Tabs-panel--description",
		func(e *colly.HTMLElement) {
			desc := strings.TrimSpace(e.Text)

			idxAny := e.Request.Ctx.Get("index")
			idx, _ := strconv.Atoi(idxAny)

			if desc == "" {
				s.logger.Warn("empty description", zap.String("url", e.Request.URL.String()))
				return
			}
			s.mu.Lock()
			s.products[idx].Description = desc
			s.mu.Unlock()
		},
	)

	for i := range s.products {
		if s.products[i].Link == "" {
			continue
		}

		ctx := colly.NewContext()
		ctx.Put("index", strconv.Itoa(i))

		if err := c.Request(
			"GET",
			s.products[i].Link,
			nil,
			ctx,
			nil,
		); err != nil {
			s.logger.Warn(
				"Failed to visit product",
				zap.String("url", s.products[i].Link),
				zap.Error(err),
			)
		}
	}

	c.Wait()
	return nil
}
