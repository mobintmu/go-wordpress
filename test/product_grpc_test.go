package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	productv1 "go-wordpress/api/proto/product/v1"
	"go-wordpress/internal/auth"
	"go-wordpress/internal/config"
	"go-wordpress/internal/storage/sql/sqlc"
)

func TestProductGRPC(t *testing.T) {
	WithHttpTestServer(t, func() {
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		base := fmt.Sprintf("http://%s:%d/api/v1/admin", cfg.HTTPAddress, cfg.HTTPPort)

		token, err := auth.GenerateToken(cfg, "admin-123")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Create prerequisites.
		var website sqlc.Website
		adminCreateWebsite(t, &website, base+"/websites", token)
		defer adminDeleteWebsite(t, website, base+"/websites", token)

		var category sqlc.Category
		adminCreateCategory(t, &category, base+"/categories", token, website.ID)
		defer adminDeleteCategory(t, category, base+"/categories", token)

		var product sqlc.Product
		adminCreateProduct(t, &product, base+"/products", token, website.ID, category.ID)
		defer adminDeleteProduct(t, product, base+"/products", token)

		// Connect to gRPC server.
		grpcAddr := cfg.HTTPAddress + ":" + strconv.Itoa(cfg.GRPCPort)
		conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("Failed to connect to gRPC server: %v", err)
		}
		defer conn.Close()

		client := productv1.NewProductServiceClient(conn)

		grpcGetProductByID(t, product, client)
	})
}

func grpcGetProductByID(t *testing.T, product sqlc.Product, client productv1.ProductServiceClient) {
	t.Run("Get Product By ID (gRPC)", func(t *testing.T) {
		resp, err := client.GetProductByID(context.Background(), &productv1.ProductRequest{
			Id: product.ID,
		})
		if err != nil {
			t.Fatalf("gRPC GetProductByID failed: %v", err)
		}

		assert.NotNil(t, resp)
		assert.Equal(t, product.ID, resp.Id)
		t.Logf("gRPC retrieved product ID: %d", resp.Id)
	})
}
