package app

import (
	categoryController "go-wordpress/internal/category/controller"
	categoryService "go-wordpress/internal/category/service"
	"go-wordpress/internal/config"
	configsController "go-wordpress/internal/configs/controller"
	configsService "go-wordpress/internal/configs/service"
	"go-wordpress/internal/health"
	"go-wordpress/internal/poller"
	"go-wordpress/internal/poller/dispatcher"
	productController "go-wordpress/internal/product/controller"
	productService "go-wordpress/internal/product/service"
	websiteController "go-wordpress/internal/website/controller"
	websiteService "go-wordpress/internal/website/service"

	"go-wordpress/internal/server"
	"go-wordpress/internal/storage/cache"
	"go-wordpress/internal/storage/sql"
	"go-wordpress/internal/storage/sql/migrate"
	"go-wordpress/internal/storage/sql/sqlc"
	"go-wordpress/pkg/logger"

	"go.uber.org/fx"
)

func NewApp() *fx.App {
	return fx.New(
		fx.Provide(
			logger.NewLogger,
			config.NewConfig,
			sql.InitialDB,
			//server
			health.New,
			server.NewGinEngine,
			server.CreateHTTPServer,
			server.CreateGRPCServer,
			//db
			migrate.NewRunner, // migration runner
			migrate.NewSeeder, // seeder for initial data
			sqlc.New,
			//cache
			cache.NewClient,
			cache.NewCacheStore,
			//controller
			productController.NewAdminProduct,
			websiteController.NewAdminWebsite,
			categoryController.NewAdminCategory,
			productController.NewGRPC,
			configsController.NewAdminConfig,
			//service
			productService.New,
			websiteService.New,
			categoryService.New,
			configsService.New,
			// dispatcher
			dispatcher.New,
			poller.New,
		),
		fx.Invoke(
			// dispatcher
			dispatcher.RegisterServices,
			dispatcher.RegisterLifecycle,

			// poller
			poller.RegisterLifecycle,

			//server
			server.RegisterRoutes,
			server.StartHTTPServer,
			server.StartGRPCServer,
			//migration
			migrate.RunMigrations,
			migrate.RunSeeder,
			//life cycle
			logger.RegisterLoggerLifecycle,
			server.GRPCLifeCycle,
		),
	)
}
