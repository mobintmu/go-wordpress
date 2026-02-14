package crawler_test

import (
	"go-wordpress/internal/crawler/cleanup"
	"go-wordpress/internal/crawler/description"
	"go-wordpress/internal/crawler/domain"
	"go-wordpress/internal/crawler/product"
	"go-wordpress/internal/crawler/website"
	"testing"

	"go.uber.org/zap"
)

func TestWebsites(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	webs := website.New(logger)
	if len(webs) == 0 {
		t.Fatalf("expected at least 1 website, got 0")
	}
	web := webs[0]
	pr := product.New(logger, &web)
	if err := pr.Run(); err != nil {
		t.Fatalf("failed to run service: %v", err)
	}

	cleanCategories := pr.CleanedProducts()
	if len(cleanCategories) == 0 {
		t.Fatalf("expected at least 1 category, got 0")
	}

	// Check first category
	cleanedProducts := cleanCategories[0].CleanedProducts
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

func TestProducts(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	webs := website.NewWithCategories(logger)
	if len(webs) == 0 {
		t.Fatalf("expected at least 1 website, got 0")
	}

	web := webs[0]

	// Ensure we don't crash if there are fewer than 10 categories
	limit := 3
	if len(webs[0].Categories) < limit {
		limit = len(webs[0].Categories)
	}
	web.Categories = web.Categories[:limit]

	web.Categories[0] = domain.Category{
		Name:            "آجیل شب یلدا",
		Link:            "https://hosseinibrothers.ir/آجیل/آجیل-های-مخلوط/آجیل-شب-یلدا/",
		Products:        []domain.Product{},
		CleanedProducts: []domain.ProductClean{},
	}
	pr := product.New(logger, &web)
	if err := pr.Run(); err != nil {
		t.Fatalf("failed to run service: %v", err)
	}

	cleanCategories := pr.CleanedProducts()
	if len(cleanCategories) == 0 {
		t.Fatalf("expected at least 1 category, got 0")
	}

	// Check first category
	cleanedProducts := cleanCategories[0].CleanedProducts
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

func TestDescription(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	webs := website.NewWithCategories(logger)
	if len(webs) == 0 {
		t.Fatalf("expected at least 1 website, got 0")
	}

	web := webs[0]
	web.Categories[0] = domain.Category{
		Name:            "آجیل شب یلدا",
		Link:            "https://hosseinibrothers.ir/آجیل/آجیل-های-مخلوط/آجیل-شب-یلدا/",
		Products:        []domain.Product{},
		CleanedProducts: []domain.ProductClean{},
	}
	web.Categories[1] = domain.Category{
		Name: "آجیل",
		Link: "https://hosseinibrothers.ir/آجیل/",
		Products: []domain.Product{
			{
				Title:       "کدو گوشتی شور",
				Price:       "748,000 تومان",
				Link:        "https://hosseinibrothers.ir/تخمه-کدو/کدو-گوشتی-ایرانی",
				Image:       "https://hosseinibrothers.ir/868-medium_default/کدو-گوشتی-ایرانی.jpg",
				Description: "",
			},
			{
				Title:       "آجیل ممتاز",
				Price:       "1,480,000 تومان",
				Link:        "https://hosseinibrothers.ir/آجیل-های-مخلوط/آجیل-ممتاز",
				Image:       "https://hosseinibrothers.ir/3006-medium_default/آجیل-ممتاز.jpg",
				Description: "",
			},
			{Title: "آجیل ممتاز", Price: "1,480,000 تومان", Link: "https://hosseinibrothers.ir/آجیل-های-مخلوط/آجیل-ممتاز", Image: "https://hosseinibrothers.ir/3006-medium_default/آجیل-ممتاز.jpg", Description: ""},
			{Title: "آجیل پذیرایی", Price: "2,035,000 تومان", Link: "https://hosseinibrothers.ir/آجیل-های-مخلوط/آجیل-پذیرایی", Image: "https://hosseinibrothers.ir/3000-medium_default/آجیل-پذیرایی.jpg", Description: ""},
			{Title: "نخود دو آتشه شور", Price: "506,000 تومان", Link: "https://hosseinibrothers.ir/نخود/نخود-دو-آتشه-شور", Image: "https://hosseinibrothers.ir/1052-medium_default/نخود-دو-آتشه-شور.jpg", Description: ""},
		},
		CleanedProducts: []domain.ProductClean{},
	}
	web.Categories[2] = domain.Category{
		Name: "پسته",
		Link: "https://hosseinibrothers.ir/آجیل/پسته/",
		Products: []domain.Product{
			{Title: "پسته احمد آقایی نمکی", Price: "2,236,000 تومان", Link: "https://hosseinibrothers.ir/پسته/پسته-نمکی", Image: "https://hosseinibrothers.ir/906-medium_default/پسته-نمکی.jpg", Description: ""},
			{Title: "پسته اکبری", Price: "2,067,000 تومان", Link: "https://hosseinibrothers.ir/پسته-اکبری/پسته-اکبری-ممتاز", Image: "https://hosseinibrothers.ir/3158-medium_default/پسته-اکبری-ممتاز.jpg", Description: ""},
			{Title: "پسته فندقی شور", Price: "", Link: "https://hosseinibrothers.ir/سایر-پسته-ها/پسته-فندقی-شور", Image: "https://hosseinibrothers.ir/840-medium_default/پسته-فندقی-شور.jpg", Description: ""},
			{Title: "پسته احمد آقایی زعفرانی", Price: "2,248,500 تومان", Link: "https://hosseinibrothers.ir/پسته-احمد-آقایی/پسته-احمد-آقایی-زعفرانی", Image: "https://hosseinibrothers.ir/904-medium_default/پسته-احمد-آقایی-زعفرانی.jpg", Description: ""},
			{Title: "پسته اکبری دستچین خام", Price: "2,430,000 تومان", Link: "https://hosseinibrothers.ir/پسته-اکبری/پسته-اکبری-دستچین-خام", Image: "https://hosseinibrothers.ir/917-medium_default/پسته-اکبری-دستچین-خام.jpg", Description: ""},
		},
		CleanedProducts: []domain.ProductClean{},
	}

	desc := description.New(logger, &web)
	w, err := desc.FetchDescriptions(&web)
	if err != nil {
		t.Fatalf("failed to fetch descriptions: %v", err)
	}
	if w == nil {
		t.Fatalf("expected web with descriptions, got nil")
	}
	cleanup.Cleans(w.Categories)
	if w.Categories[1].CleanedProducts[0].Description == "" {
		t.Errorf("expected description for first product, got empty string")
	}

}
