package crawler

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Service struct {
	products        []Product
	cleanedProducts []Product
}

func New() *Service {
	return &Service{
		products: []Product{},
	}
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
		products = append(products, p)
	})

	err := c.Visit("https://badomjip.com/product-category/%d9%be%da%a9%db%8c%d8%ac%d9%87%d8%a7/")
	if err != nil {
		return err
	}

	s.products = products
	s.cleanUP()
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
			Title: title,
			Price: strconv.Itoa(priceInt), // or change Product.Price to int if you prefer
			Link:  p.Link,
			Image: p.Image,
		})
	}
	s.cleanedProducts = cleaned
}

func (s *Service) CleanedProducts() []Product { return s.cleanedProducts }
