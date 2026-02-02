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
	web    *Website
	logger *zap.Logger
	mu     sync.Mutex
}

func New(logger *zap.Logger, web *Website) *Service {
	return &Service{
		logger: logger,
		web:    web,
		mu:     sync.Mutex{},
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

	for index, cat := range s.web.Categories {
		c := colly.NewCollector(
			colly.AllowedDomains(s.web.Domain),
		)

		products := []Product{}

		// Use configured selectors from web.ProductList
		c.OnHTML(s.web.ProductList.Item, func(e *colly.HTMLElement) {
			p := Product{
				Title: e.ChildText(s.web.ProductList.Fields["title"].Selector),
				Price: e.ChildText(s.web.ProductList.Fields["price"].Selector),
				Link:  e.ChildAttr(s.web.ProductList.Fields["link"].Selector, s.web.ProductList.Fields["link"].Attr),
				Image: e.ChildAttr(s.web.ProductList.Fields["image"].Selector, s.web.ProductList.Fields["image"].Attr),
			}

			if p.Link != "" {
				products = append(products, p)
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

		err := c.Visit(cat.Link)
		if err != nil {
			return err
		}

		s.web.Categories[index].Products = products
	}
	return nil
}

func (s *Service) cleanUP() {

	for index, cat := range s.web.Categories {
		cleaned := make([]ProductClean, 0, len(cat.Products))

		// regex to strip HTML tags
		reTags := regexp.MustCompile(`<[^>]*>`)

		for _, p := range cat.Products {
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

			priceStr = convertPersianDigits(priceStr)

			var priceInt int
			if priceStr != "" {
				if val, err := strconv.Atoi(priceStr); err == nil {
					priceInt = val
				}
			}

			// --- Build cleaned product ---
			cleaned = append(cleaned, ProductClean{
				Title:       title,
				Price:       int64(priceInt),
				Link:        p.Link,
				Image:       p.Image,
				Description: p.Description,
			})
		}
		s.web.Categories[index].CleanedProducts = cleaned
	}
}

func convertPersianDigits(str string) string {
	persianDigits := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	for i, digit := range persianDigits {
		str = strings.ReplaceAll(str, digit, strconv.Itoa(i))
	}
	return str
}

func (s *Service) CleanedProducts() []Category { return s.web.Categories }

func (s *Service) FetchDescriptions() error {

	for index, cat := range s.web.Categories {
		c := colly.NewCollector(
			colly.AllowedDomains(s.web.Domain),
			colly.Async(true), // ENABLE ASYNC
		)

		// Control parallelism
		c.Limit(&colly.LimitRule{
			DomainGlob:  s.web.Domain,
			Parallelism: 4, // number of concurrent requests
			Delay:       800 * time.Millisecond,
		})

		// Extract description from product page
		c.OnHTML(
			s.web.ProductDetail.Description.Selector,
			func(e *colly.HTMLElement) {
				desc := strings.TrimSpace(e.Text)

				idxAny := e.Request.Ctx.Get("index")
				idx, _ := strconv.Atoi(idxAny)

				if desc == "" {
					s.logger.Warn("empty description", zap.String("url", e.Request.URL.String()))
					return
				}
				s.mu.Lock()
				s.web.Categories[index].Products[idx].Description = desc
				s.mu.Unlock()
			},
		)

		for i := range cat.Products {
			if s.web.Categories[index].Products[i].Link == "" {
				continue
			}

			ctx := colly.NewContext()
			ctx.Put("index", strconv.Itoa(i))

			if err := c.Request(
				"GET",
				s.web.Categories[index].Products[i].Link,
				nil,
				ctx,
				nil,
			); err != nil {
				s.logger.Warn(
					"Failed to visit product",
					zap.String("url", s.web.Categories[index].Products[i].Link),
					zap.Error(err),
				)
			}
		}

		c.Wait()

	}
	return nil
}
