package test

import (
	"context"
	"os"
	"testing"

	app "go-wordpress/internal/app"
	"go-wordpress/internal/config"
	"go-wordpress/internal/crawler/category"
	"go-wordpress/internal/crawler/website"
	"go-wordpress/internal/storage/sql/sqlc"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type crawlerDeps struct {
	websites *website.Service
	logger   *zap.Logger
	query    *sqlc.Queries
}

func setupCrawler(t *testing.T) *crawlerDeps {
	t.Helper()

	// Move to project root BEFORE loading env
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// Ensure we return to /test after the test finishes
	t.Cleanup(func() { os.Chdir("test") })

	// Set test environment
	os.Setenv("APP_ENV", "test")
	config.LoadEnv()

	deps := &crawlerDeps{}

	// Build full FX app
	a := fx.New(
		app.AppOptions(),
		// The crawler website service is not part of AppOptions.
		fx.Provide(website.New),
		fx.Populate(&deps.websites, &deps.logger, &deps.query),
	)

	// Start app
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("failed to start app: %v", err)
	}

	// Stop app after test
	t.Cleanup(func() { a.Stop(context.Background()) })

	if deps.websites == nil {
		t.Fatal("website service is nil")
	}
	if len(deps.websites.Websites) == 0 {
		t.Fatal("no websites found in website service")
	}

	return deps
}

func TestCategoryDiscoveryAndSave(t *testing.T) {
	deps := setupCrawler(t)
	ctx := context.Background()

	for i := range deps.websites.Websites {
		// Use a pointer: category.Service stores results on the website.
		web := &deps.websites.Websites[i]

		catSvc, err := category.New(deps.logger, deps.query, web)
		if err != nil {
			t.Fatalf("failed to create category service for %s: %v", web.Name, err)
		}

		if err := catSvc.Discover(); err != nil {
			t.Fatalf("failed to discover categories for %s: %v", web.Name, err)
		}

		if err := catSvc.Save(ctx); err != nil {
			t.Fatalf("failed to save categories for %s: %v", web.Name, err)
		}

		catSvc.Print()
		t.Logf("Discovered and saved %d categories for website: %s", len(web.Categories), web.Name)
	}
}

func TestProductDescriptionFetch(t *testing.T) {
	deps := setupCrawler(t)

	web := &deps.websites.Websites[0]

	catSvc, err := category.New(deps.logger, deps.query, web)
	if err != nil {
		t.Fatalf("failed to create category service for %s: %v", web.Name, err)
	}

	if err := catSvc.Load(context.Background()); err != nil {
		t.Fatalf("failed to load category: %v", err)
	}
}
