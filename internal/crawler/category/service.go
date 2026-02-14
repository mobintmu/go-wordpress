package category

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

// Service helps find all categories on a website
type Service struct {
	logger     *zap.Logger
	domain     string
	startURL   string
	Categories []Category
}

// Category represents a category found on the website
type Category struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// New creates a new category discovery service
func New(logger *zap.Logger, domain string, startURL string) *Service {
	return &Service{
		logger:     logger,
		domain:     domain,
		startURL:   startURL,
		Categories: []Category{},
	}
}

// Discover finds all categories from the homepage
func (cd *Service) Discover() error {

	c := colly.NewCollector(
		colly.AllowedDomains(cd.domain),
	)

	var categories []Category

	// Helper function to clean text
	cleanText := func(text string) string {
		// strings.TrimSpace handles \n, \t, and whitespace from both ends
		return strings.TrimSpace(text)
	}

	// Try to find categories in common navigation structures
	// Adjust selectors based on actual website structure
	c.OnHTML("nav a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		text := cleanText(e.Text)

		// Filter out irrelevant links
		if link != "" && text != "" && isValidLink(link, cd.domain) {
			categories = append(categories, Category{
				Name: text,
				Link: link,
			})
			cd.logger.Debug("Found category",
				zap.String("name", text),
				zap.String("link", link),
			)
		}
	})

	// Also try product category selectors
	c.OnHTML("a.category-link, div.product-category a, .category-menu a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		text := cleanText(e.Text)

		if link != "" && text != "" && isValidLink(link, cd.domain) {
			// Check if already exists
			exists := false
			for _, cat := range categories {
				if cat.Link == link {
					exists = true
					break
				}
			}

			if !exists {
				categories = append(categories, Category{
					Name: text,
					Link: link,
				})
				cd.logger.Debug("Found category",
					zap.String("name", text),
					zap.String("link", link),
				)
			}
		}
	})

	// Add this before calling c.Visit()
	c.OnError(func(r *colly.Response, err error) {
		cd.logger.Warn("Request failed, skipping link",
			zap.String("url", r.Request.URL.String()),
			zap.Int("status_code", r.StatusCode),
			zap.Error(err),
		)
	})

	err := c.Visit(cd.startURL)
	if err != nil {
		cd.logger.Warn("Failed to visit start URL", zap.Error(err))
		return err
	}

	cd.logger.Info("Category discovery completed",
		zap.Int("count", len(categories)),
	)

	cd.Categories = categories
	return nil
}

// isValidLink checks if a link is valid and relevant
func isValidLink(link string, baseDomain string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	// 1. Domain Check (Removes different domains)
	if u.Host != "" && u.Host != baseDomain {
		return false
	}

	path := strings.ToLower(u.Path)

	// 2. Standard Exclusions
	invalidPatterns := []string{
		"/login", "/register", "/account", "/cart",
		"/checkout", "/admin", "/blog", "/contact-us", "/about-us",
		"/content",
	}
	for _, pattern := range invalidPatterns {
		if strings.HasPrefix(path, pattern) {
			return false
		}
	}

	// 3. Specific Exclusions for your request:

	// Remove specific products but keep /product-category/
	if strings.HasPrefix(path, "/product/") {
		return false
	}

	// Remove the Persian Registration Form link
	// %d9%81%d8%b1%d9%85 is the encoded version of "فرم" (Form)
	if strings.Contains(path, "%d9%81%d8%b1%d9%85") || strings.Contains(path, "فرم-ثبت-نام") {
		return false
	}

	// 4. Schemes and Fragments
	scheme := strings.ToLower(u.Scheme)
	if scheme == "javascript" || scheme == "mailto" || scheme == "tel" || u.Fragment != "" {
		return false
	}

	return path != "" && path != "/"
}

// SaveToFile exports discovered categories to a configuration file
func (cd *Service) SaveToFile() error {
	filePath := fmt.Sprintf("config/%s_categories.json", cd.domain)
	config := map[string]interface{}{
		"domain":     cd.domain,
		"categories": cd.Categories,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0o644)
	if err != nil {
		return err
	}

	cd.logger.Info("Categories saved to file",
		zap.String("path", filePath),
		zap.Int("count", len(cd.Categories)),
	)

	return nil
}

func (cd *Service) LoadFromFile() error {
	filePath := fmt.Sprintf("config/%s_categories.json", cd.domain)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config struct {
		Domain     string     `json:"domain"`
		Categories []Category `json:"categories"`
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	cd.Categories = config.Categories
	cd.logger.Info("Categories loaded from file",
		zap.String("path", filePath),
		zap.Int("count", len(cd.Categories)),
	)

	return nil
}

// Print displays categories in a readable format
func (cd *Service) Print(categories []Category) {
	fmt.Println("\n========== Discovered Categories ==========")
	for i, cat := range categories {
		fmt.Printf("%d. %s\n   URL: %s\n\n", i+1, cat.Name, cat.Link)
	}
	fmt.Printf("Total categories found: %d\n", len(categories))
}
