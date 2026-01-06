package crawler_test

import (
	"go-wordpress/internal/crawler"
	"testing"
)

func TestListProductsIntegration(t *testing.T) {
	svc := crawler.New()

	err := svc.ListProducts()
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	products := svc.CleanedProducts() // expose getter or access field directly if exported

	if len(products) == 0 {
		t.Fatalf("expected at least 1 product, got 0")
	}

	// Basic sanity checks on first product
	p := products[0]

	if p.Title == "" {
		t.Errorf("product title should not be empty")
	}
	if p.Price == "" {
		t.Errorf("product price should not be empty")
	}
	if p.Link == "" {
		t.Errorf("product link should not be empty")
	}
	if p.Image == "" {
		t.Errorf("product image should not be empty")
	}
}
