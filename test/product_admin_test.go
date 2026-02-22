package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-wordpress/internal/auth"
	"go-wordpress/internal/config"
	"go-wordpress/internal/product/dto"
	"go-wordpress/internal/storage/sql/sqlc"
	"io"
	"net/http"
	"testing"
)

func TestProductsAdmin(t *testing.T) {
	WithHttpTestServer(t, func() {
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		base := fmt.Sprintf("http://%s:%d/api/v1/admin", cfg.HTTPAddress, cfg.HTTPPort)
		productAddr := base + "/products"

		token, err := auth.GenerateToken(cfg, "admin-123")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Create prerequisite website and category to satisfy FK constraints.
		var website sqlc.Website
		adminCreateWebsite(t, &website, base+"/websites", token)
		defer adminDeleteWebsite(t, website, base+"/websites", token)

		var category sqlc.Category
		adminCreateCategory(t, &category, base+"/categories", token, website.ID)
		defer adminDeleteCategory(t, category, base+"/categories", token)

		var product sqlc.Product

		adminUnAuthorizedCreateProduct(t, productAddr)
		adminUnAuthorizedListProduct(t, productAddr)
		adminCreateProduct(t, &product, productAddr, token, website.ID, category.ID)
		adminListProduct(t, product, productAddr, token)
		adminGetProductByID(t, product, productAddr, token)
		adminUpdateProduct(t, product, productAddr, token, website.ID, category.ID)
		adminDeleteProduct(t, product, productAddr, token)
		adminVerifyProductDeleted(t, product, productAddr, token)
	})
}

func testCreateProductParams(websiteID, categoryID int32) sqlc.CreateProductParams {
	return sqlc.CreateProductParams{
		WebsiteID:   websiteID,
		CategoryID:  categoryID,
		Title:       "Test Product",
		Price:       1000,
		Link:        "https://example.com/test-product",
		Image:       sql.NullString{String: "https://example.com/image.jpg", Valid: true},
		Description: sql.NullString{String: "This is a test product", Valid: true},
		Status:      sqlc.EntityStatusActive,
	}
}

func testUpdateProductParams(id, websiteID, categoryID int32) sqlc.UpdateProductParams {
	return sqlc.UpdateProductParams{
		ID:          id,
		WebsiteID:   websiteID,
		CategoryID:  categoryID,
		Title:       "Updated Product",
		Price:       1500,
		Link:        "https://example.com/updated-product",
		Image:       sql.NullString{String: "https://example.com/updated-image.jpg", Valid: true},
		Description: sql.NullString{String: "Updated description", Valid: true},
		Status:      sqlc.EntityStatusInactive,
	}
}

func adminUnAuthorizedCreateProduct(t *testing.T, addr string) {
	t.Run("Unauthorized Create Product", func(t *testing.T) {
		body, err := json.Marshal(testCreateProductParams(1, 1))
		if err != nil {
			t.Fatalf("Failed to marshal product: %v", err)
		}
		resp, err := http.Post(addr, ApplicationJsonHeader, bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
			responseBody, _ := io.ReadAll(resp.Body)
			t.Logf(ResponseBodyMessage, string(responseBody))
		}
	})
}

func adminUnAuthorizedListProduct(t *testing.T, addr string) {
	t.Run("Unauthorized List Products", func(t *testing.T) {
		resp, err := http.Get(addr)
		if err != nil {
			t.Fatalf(FailedToSendGetMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
			responseBody, _ := io.ReadAll(resp.Body)
			t.Logf(ResponseBodyMessage, string(responseBody))
		}
	})
}

func adminCreateProduct(t *testing.T, product *sqlc.Product, addr string, token string, websiteID, categoryID int32) {
	t.Run("Create Product", func(t *testing.T) {
		body, err := json.Marshal(testCreateProductParams(websiteID, categoryID))
		if err != nil {
			t.Fatalf("Failed to marshal product: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, addr, bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", ApplicationJsonHeader)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			responseBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 201 Created, got %d: %s", resp.StatusCode, string(responseBody))
		}

		if err := json.NewDecoder(resp.Body).Decode(product); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if product.ID == 0 {
			t.Fatalf("Expected a non-zero product ID after creation")
		}
		if product.Title != testCreateProductParams(websiteID, categoryID).Title {
			t.Errorf("Expected title %q, got %q", testCreateProductParams(websiteID, categoryID).Title, product.Title)
		}
		if product.WebsiteID != websiteID {
			t.Errorf("Expected website ID %d, got %d", websiteID, product.WebsiteID)
		}
		if product.CategoryID != categoryID {
			t.Errorf("Expected category ID %d, got %d", categoryID, product.CategoryID)
		}
		if product.Status != sqlc.EntityStatusActive {
			t.Errorf("Expected status %q, got %q", sqlc.EntityStatusActive, product.Status)
		}
		t.Logf("Created product: %+v", product)
	})
}

func adminListProduct(t *testing.T, product sqlc.Product, addr string, token string) {
	t.Run("List Products", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, addr, nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf(FailedToSendGetMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseBody, _ := io.ReadAll(resp.Body)
			t.Fatalf(ExpectedStatus200OKGotMessage+": %s", resp.StatusCode, string(responseBody))
		}

		var products dto.ProductsResponse
		if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if len(products) == 0 {
			t.Fatalf("Expected at least one product, got 0")
		}

		for i, p := range products {
			if p.ID == product.ID {
				t.Logf("Found created product at index %d: %+v", i, p)
				return
			}
		}
		t.Errorf("Created product (ID=%d) not found in list", product.ID)
	})
}

func adminGetProductByID(t *testing.T, product sqlc.Product, addr string, token string) {
	t.Run("Get Product By ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, product.ID), nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf(FailedToSendGetMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(responseBody))
		}

		var fetched sqlc.Product
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if fetched.ID != product.ID {
			t.Errorf("Expected product ID %d, got %d", product.ID, fetched.ID)
		}
		if fetched.Title != product.Title {
			t.Errorf("Expected product title %q, got %q", product.Title, fetched.Title)
		}
		if fetched.WebsiteID != product.WebsiteID {
			t.Errorf("Expected website ID %d, got %d", product.WebsiteID, fetched.WebsiteID)
		}
		if fetched.CategoryID != product.CategoryID {
			t.Errorf("Expected category ID %d, got %d", product.CategoryID, fetched.CategoryID)
		}
	})
}

func adminUpdateProduct(t *testing.T, product sqlc.Product, addr string, token string, websiteID, categoryID int32) {
	t.Run("Update Product", func(t *testing.T) {
		updateReq := testUpdateProductParams(product.ID, websiteID, categoryID)

		body, err := json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", addr, product.ID), bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create PUT request: %v", err)
		}
		req.Header.Set("Content-Type", ApplicationJsonHeader)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("Failed to send PUT request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200 OK, got %d: %s", resp.StatusCode, string(responseBody))
		}

		var updated sqlc.Product
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if updated.Title != updateReq.Title {
			t.Errorf("Expected title %q, got %q", updateReq.Title, updated.Title)
		}
		if updated.Price != updateReq.Price {
			t.Errorf("Expected price %d, got %d", updateReq.Price, updated.Price)
		}
		if updated.CategoryID != updateReq.CategoryID {
			t.Errorf("Expected category ID %d, got %d", updateReq.CategoryID, updated.CategoryID)
		}
		if updated.Status != updateReq.Status {
			t.Errorf("Expected status %q, got %q", updateReq.Status, updated.Status)
		}
		if updated.Link != updateReq.Link {
			t.Errorf("Expected link %q, got %q", updateReq.Link, updated.Link)
		}
	})
}

func adminDeleteProduct(t *testing.T, product sqlc.Product, addr string, token string) {
	t.Run("Delete Product", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", addr, product.ID), nil)
		if err != nil {
			t.Fatalf("Failed to create DELETE request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("Failed to send DELETE request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			responseBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 204 No Content, got %d: %s", resp.StatusCode, string(responseBody))
		}
	})
}

func adminVerifyProductDeleted(t *testing.T, product sqlc.Product, addr string, token string) {
	t.Run("Verify Product Deleted", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, product.ID), nil)
		if err != nil {
			t.Fatalf("Failed to create GET request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf(FailedToSendGetMessage, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected error status (400 or 500) after deletion, got %d", resp.StatusCode)
		}
	})
}
