package category

import (
	"context"
	"fmt"
	"go-wordpress/internal/crawler/domain"
	"go-wordpress/internal/storage/sql/sqlc"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

type Service struct {
	logger   *zap.Logger
	domain   string
	startURL string
	query    *sqlc.Queries
	web      *domain.Website
}

func New(logger *zap.Logger, query *sqlc.Queries, web *domain.Website) (*Service, error) {
	if web == nil {
		return nil, fmt.Errorf("website cannot be nil")
	}
	if web.CFG.StartURL == "" {
		return nil, fmt.Errorf("website start URL cannot be empty")
	}
	if web.Domain == "" {
		return nil, fmt.Errorf("website domain cannot be empty")
	}

	return &Service{
		logger:   logger,
		domain:   web.Domain,
		startURL: web.CFG.StartURL,
		query:    query,
		web:      web,
	}, nil
}

func (s *Service) Discover() error {
	c := colly.NewCollector(
		colly.AllowedDomains(s.domain),
	)

	categories := make(map[string]domain.Category)

	clean := func(t string) string { 
		return strings.TrimSpace(t) 
	}

	handleCategory := func(link, text string) {
		if link == "" || text == "" {
			return
		}
		if !isValidLink(link, s.domain) {
			return
		}
		if _, exists := categories[link]; exists {
			return
		}

		categories[link] = domain.Category{Name: text, Link: link}

		s.logger.Debug("Found category",
			zap.String("name", text),
			zap.String("link", link),
		)
	}

	// Generic nav links
	c.OnHTML("nav a[href]", func(e *colly.HTMLElement) {
		handleCategory(e.Request.AbsoluteURL(e.Attr("href")), clean(e.Text))
	})

	// WooCommerce category selectors
	c.OnHTML("a.category-link, div.product-category a, .category-menu a", func(e *colly.HTMLElement) {
		handleCategory(e.Request.AbsoluteURL(e.Attr("href")), clean(e.Text))
	})

	c.OnError(func(r *colly.Response, err error) {
		s.logger.Warn("Request failed",
			zap.String("url", r.Request.URL.String()),
			zap.Int("status_code", r.StatusCode),
			zap.Error(err),
		)
	})

	if err := c.Visit(s.startURL); err != nil {
		s.logger.Warn("Failed to visit start URL", zap.Error(err))
		return err
	}

	// Convert map → slice
	s.web.Categories = make([]domain.Category, 0, len(categories))
	for _, cat := range categories {
		s.web.Categories = append(s.web.Categories, cat)
	}

	s.logger.Info("Category discovery completed",
		zap.Int("count", len(s.web.Categories)),
	)

	return nil
}

func isValidLink(link, baseDomain string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	// External domain
	if u.Host != "" && u.Host != baseDomain {
		return false
	}

	path := strings.ToLower(u.Path)
	if path == "" || path == "/" {
		return false
	}

	invalid := []string{
		"/login", "/register", "/account", "/cart",
		"/checkout", "/admin", "/blog", "/contact-us",
		"/about-us", "/content",
	}
	for _, p := range invalid {
		if strings.HasPrefix(path, p) {
			return false
		}
	}

	// Exclude product pages
	if strings.HasPrefix(path, "/product/") {
		return false
	}

	// Persian registration form
	if strings.Contains(path, "%d9%81%d8%b1%d9%85") || strings.Contains(path, "فرم-ثبت-نام") {
		return false
	}

	// Schemes
	switch strings.ToLower(u.Scheme) {
	case "javascript", "mailto", "tel":
		return false
	}

	if u.Fragment != "" {
		return false
	}

	return true
}

func (s *Service) Save(ctx context.Context) error {
	if s.web.ID == 0 {
		return fmt.Errorf("website ID is not set")
	}
	if len(s.web.Categories) == 0 {
		return fmt.Errorf("no categories to save for website ID %d", s.web.ID)
	}

	for _, cat := range s.web.Categories {
		_, err := s.query.UpsertCategory(ctx, sqlc.UpsertCategoryParams{
			WebsiteID: s.web.ID,
			Name:      cat.Name,
			Link:      cat.Link,
			Status:    sqlc.EntityStatusActive,
		})
		if err != nil {
			s.logger.Error("Failed to upsert category",
				zap.String("link", cat.Link),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

func (s *Service) Load(ctx context.Context) error {
	if s.web.ID == 0 {
		return fmt.Errorf("website ID is not set")
	}

	items, err := s.query.ListActiveCategoriesByWebsiteID(ctx, sqlc.ListActiveCategoriesByWebsiteIDParams{
		WebsiteID: s.web.ID,
		Limit:     domain.DefaultLimit,
		Offset:    domain.DefaultOffset,
	})
	if err != nil {
		s.logger.Error("Failed to list active categories", zap.Error(err))
		return err
	}

	s.web.Categories = make([]domain.Category, len(items))
	for i, item := range items {
		s.web.Categories[i] = domain.Category{
			Name: item.Name,
			Link: item.Link,
		}
	}

	return nil
}

func (s *Service) Print() {
	fmt.Println("\n========== Discovered Categories ==========")
	for i, cat := range s.web.Categories {
		fmt.Printf("%d. %s\n   URL: %s\n\n", i+1, cat.Name, cat.Link)
	}
	fmt.Printf("Total categories found: %d\n", len(s.web.Categories))
}
