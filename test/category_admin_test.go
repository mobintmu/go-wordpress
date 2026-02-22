package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-wordpress/internal/auth"
	"go-wordpress/internal/category/dto"
	"go-wordpress/internal/config"
	"go-wordpress/internal/storage/sql/sqlc"
	"io"
	"net/http"
	"testing"
)

func TestCategoriesAdmin(t *testing.T) {
	WithHttpTestServer(t, func() {
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		base := fmt.Sprintf("http://%s:%d/api/v1/admin", cfg.HTTPAddress, cfg.HTTPPort)
		categoryAddr := base + "/categories"

		token, err := auth.GenerateToken(cfg, "admin-123")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Categories have a FK on website_id, so create a website first.
		var website sqlc.Website
		adminCreateWebsite(t, &website, base+"/websites", token)
		defer adminDeleteWebsite(t, website, base+"/websites", token)

		var category sqlc.Category

		adminUnAuthorizedCreateCategory(t, categoryAddr)
		adminUnAuthorizedListCategory(t, categoryAddr)
		adminCreateCategory(t, &category, categoryAddr, token, website.ID)
		adminListCategory(t, category, categoryAddr, token)
		adminGetCategoryByID(t, category, categoryAddr, token)
		adminUpdateCategory(t, category, categoryAddr, token, website.ID)
		adminDeleteCategory(t, category, categoryAddr, token)
		adminVerifyCategoryDeleted(t, category, categoryAddr, token)
	})
}

func testCreateCategoryParams(websiteID int32) sqlc.CreateCategoryParams {
	return sqlc.CreateCategoryParams{
		WebsiteID: websiteID,
		Name:      "Test Category",
		Link:      "https://example.com/test-category",
	}
}

func testUpdateCategoryParams(id, websiteID int32) sqlc.UpdateCategoryParams {
	return sqlc.UpdateCategoryParams{
		ID:        id,
		WebsiteID: websiteID,
		Name:      "Updated Category",
		Link:      "https://example.com/updated-category",
		Status:    sqlc.EntityStatusInactive,
	}
}

func adminUnAuthorizedCreateCategory(t *testing.T, addr string) {
	t.Run("Unauthorized Create Category", func(t *testing.T) {
		body, err := json.Marshal(testCreateCategoryParams(1))
		if err != nil {
			t.Fatalf("Failed to marshal category: %v", err)
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

func adminUnAuthorizedListCategory(t *testing.T, addr string) {
	t.Run("Unauthorized List Categories", func(t *testing.T) {
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

func adminCreateCategory(t *testing.T, category *sqlc.Category, addr string, token string, websiteID int32) {
	t.Run("Create Category", func(t *testing.T) {
		body, err := json.Marshal(testCreateCategoryParams(websiteID))
		if err != nil {
			t.Fatalf("Failed to marshal category: %v", err)
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

		if err := json.NewDecoder(resp.Body).Decode(category); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if category.ID == 0 {
			t.Fatalf("Expected a non-zero category ID after creation")
		}
		if category.Name != testCreateCategoryParams(websiteID).Name {
			t.Errorf("Expected name %q, got %q", testCreateCategoryParams(websiteID).Name, category.Name)
		}
		if category.Link != testCreateCategoryParams(websiteID).Link {
			t.Errorf("Expected link %q, got %q", testCreateCategoryParams(websiteID).Link, category.Link)
		}
		if category.WebsiteID != websiteID {
			t.Errorf("Expected website ID %d, got %d", websiteID, category.WebsiteID)
		}
		t.Logf("Created category: %+v", category)
	})
}

func adminListCategory(t *testing.T, category sqlc.Category, addr string, token string) {
	t.Run("List Categories", func(t *testing.T) {
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

		var categories dto.CategoriesResponse
		if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if len(categories) == 0 {
			t.Fatalf("Expected at least one category, got 0")
		}

		for i, c := range categories {
			if c.ID == category.ID {
				t.Logf("Found created category at index %d: %+v", i, c)
				return
			}
		}
		t.Errorf("Created category (ID=%d) not found in list", category.ID)
	})
}

func adminGetCategoryByID(t *testing.T, category sqlc.Category, addr string, token string) {
	t.Run("Get Category By ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, category.ID), nil)
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

		var fetched sqlc.Category
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if fetched.ID != category.ID {
			t.Errorf("Expected category ID %d, got %d", category.ID, fetched.ID)
		}
		if fetched.Name != category.Name {
			t.Errorf("Expected category name %q, got %q", category.Name, fetched.Name)
		}
		if fetched.Link != category.Link {
			t.Errorf("Expected category link %q, got %q", category.Link, fetched.Link)
		}
		if fetched.WebsiteID != category.WebsiteID {
			t.Errorf("Expected website ID %d, got %d", category.WebsiteID, fetched.WebsiteID)
		}
	})
}

func adminUpdateCategory(t *testing.T, category sqlc.Category, addr string, token string, websiteID int32) {
	t.Run("Update Category", func(t *testing.T) {
		updateReq := testUpdateCategoryParams(category.ID, websiteID)

		body, err := json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", addr, category.ID), bytes.NewBuffer(body))
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

		var updated sqlc.Category
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if updated.Name != updateReq.Name {
			t.Errorf("Expected name %q, got %q", updateReq.Name, updated.Name)
		}
		if updated.Link != updateReq.Link {
			t.Errorf("Expected link %q, got %q", updateReq.Link, updated.Link)
		}
		if updated.Status != updateReq.Status {
			t.Errorf("Expected status %q, got %q", updateReq.Status, updated.Status)
		}
		if updated.WebsiteID != updateReq.WebsiteID {
			t.Errorf("Expected website ID %d, got %d", updateReq.WebsiteID, updated.WebsiteID)
		}
	})
}

func adminDeleteCategory(t *testing.T, category sqlc.Category, addr string, token string) {
	t.Run("Delete Category", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", addr, category.ID), nil)
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

func adminVerifyCategoryDeleted(t *testing.T, category sqlc.Category, addr string, token string) {
	t.Run("Verify Category Deleted", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, category.ID), nil)
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
