package crawler_test

import (
	"go-wordpress/internal/crawler"
	"testing"

	"go.uber.org/zap"
)

func TestBadomjip(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	web := crawler.NewWebsite("badomjip.json")
	cr := crawler.New(logger, web)
	if err := cr.Run(); err != nil {
		t.Fatalf("failed to run service: %v", err)
	}

	categories := cr.CleanedProducts()
	if len(categories) == 0 {
		t.Fatalf("expected at least 1 category, got 0")
	}

	// Check first category
	cleanedProducts := categories[0].CleanedProducts
	if len(cleanedProducts) == 0 {
		t.Fatalf("expected at least 1 product, got 0")
	}

	// Basic sanity checks on first product
	p := cleanedProducts[0]
	if p.Title == "" {
		t.Errorf("product title should not be empty")
	}
	if p.Price == 0 {
		t.Errorf("product price should not be 0")
	}
	if p.Link == "" {
		t.Errorf("product link should not be empty")
	}
	if p.Image == "" {
		t.Errorf("product image should not be empty")
	}
}

func TestHosseiniBrothers(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	web := crawler.NewWebsite("hosseinibrothers.json")
	cr := crawler.New(logger, web)

	if err := cr.Run(); err != nil {
		t.Fatalf("failed to run service: %v", err)
	}

	categories := cr.CleanedProducts()
	if len(categories) == 0 {
		t.Fatalf("expected at least 1 category, got 0")
	}

	// Check first category
	cleanedProducts := categories[0].CleanedProducts
	if len(cleanedProducts) == 0 {
		t.Fatalf("expected at least 1 product, got 0")
	}

	// Basic sanity checks on first product
	p := cleanedProducts[0]
	if p.Title == "" {
		t.Errorf("product title should not be empty")
	}
	if p.Price == 0 {
		t.Errorf("product price should not be 0")
	}
	if p.Link == "" {
		t.Errorf("product link should not be empty")
	}
	if p.Image == "" {
		t.Errorf("product image should not be empty")
	}
}
