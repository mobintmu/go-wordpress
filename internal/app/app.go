package app

import (
	"go-wordpress/internal/config"
	"go-wordpress/internal/crawler/website"
	"go-wordpress/internal/health"
	"go-wordpress/internal/poller"
	"go-wordpress/internal/poller/dispatcher"
	productController "go-wordpress/internal/product/controller"
	productService "go-wordpress/internal/product/service"
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
			sqlc.New,
			//cache
			cache.NewClient,
			cache.NewCacheStore,
			//controller
			productController.NewAdmin,
			productController.NewClient,
			productController.NewGRPC,
			//service
			productService.New,
			website.New,
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
			//life cycle
			logger.RegisterLoggerLifecycle,
			server.GRPCLifeCycle,
		),
	)
}
