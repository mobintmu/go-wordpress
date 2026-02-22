package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-wordpress/internal/auth"
	"go-wordpress/internal/config"
	"go-wordpress/internal/storage/sql/sqlc"
	"go-wordpress/internal/website/dto"
	"io"
	"net/http"
	"testing"
)

func TestWebsitesAdmin(t *testing.T) {
	WithHttpTestServer(t, func() {
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		addr := fmt.Sprintf("http://%s:%d", cfg.HTTPAddress, cfg.HTTPPort)
		addr += "/api/v1/admin/websites"

		var website sqlc.Website

		adminUnAuthorizedCreateWebsite(t, addr)
		adminUnAuthorizedListWebsite(t, addr)

		token, err := auth.GenerateToken(cfg, "admin-123")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		adminCreateWebsite(t, &website, addr, token)
		adminListWebsite(t, website, addr, token)
		adminGetWebsiteByID(t, website, addr, token)
		adminUpdateWebsite(t, website, addr, token)
		adminDeleteWebsite(t, website, addr, token)
		adminVerifyWebsiteDeleted(t, website, addr, token)
	})
}

func testCreateWebsiteParams() sqlc.CreateWebsiteParams {
	return sqlc.CreateWebsiteParams{
		Name:   "Test Website",
		Domain: "test.example.com",
	}
}

func testUpdateWebsiteParams(id int32) sqlc.UpdateWebsiteParams {
	return sqlc.UpdateWebsiteParams{
		ID:     id,
		Name:   "Updated Website",
		Domain: "updated.example.com",
		Status: sqlc.EntityStatusInactive,
	}
}

func adminUnAuthorizedCreateWebsite(t *testing.T, addr string) {
	t.Run("Unauthorized Create Website", func(t *testing.T) {
		body, err := json.Marshal(testCreateWebsiteParams())
		if err != nil {
			t.Fatalf("Failed to marshal website: %v", err)
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

func adminUnAuthorizedListWebsite(t *testing.T, addr string) {
	t.Run("Unauthorized List Websites", func(t *testing.T) {
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

func adminCreateWebsite(t *testing.T, website *sqlc.Website, addr string, token string) {
	t.Run("Create Website", func(t *testing.T) {
		body, err := json.Marshal(testCreateWebsiteParams())
		if err != nil {
			t.Fatalf("Failed to marshal website: %v", err)
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

		// CreateWebsite returns CreateWebsiteRow (id, name, domain only),
		// so decode into that then copy the fields we need for subsequent tests.
		var row sqlc.CreateWebsiteRow
		if err := json.NewDecoder(resp.Body).Decode(&row); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if row.ID == 0 {
			t.Fatalf("Expected a non-zero website ID after creation")
		}
		if row.Name != testCreateWebsiteParams().Name {
			t.Fatalf("Expected name %q, got %q", testCreateWebsiteParams().Name, row.Name)
		}
		if row.Domain != testCreateWebsiteParams().Domain {
			t.Fatalf("Expected domain %q, got %q", testCreateWebsiteParams().Domain, row.Domain)
		}

		website.ID = row.ID
		website.Name = row.Name
		website.Domain = row.Domain
		t.Logf("Created website: %+v", row)
	})
}

func adminListWebsite(t *testing.T, website sqlc.Website, addr string, token string) {
	t.Run("List Websites", func(t *testing.T) {
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

		var websites dto.WebsitesResponse
		if err := json.NewDecoder(resp.Body).Decode(&websites); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if len(websites) == 0 {
			t.Fatalf("Expected at least one website, got 0")
		}

		for i, w := range websites {
			if w.ID == website.ID {
				t.Logf("Found created website at index %d: %+v", i, w)
				return
			}
		}
		t.Errorf("Created website (ID=%d) not found in list", website.ID)
	})
}

func adminGetWebsiteByID(t *testing.T, website sqlc.Website, addr string, token string) {
	t.Run("Get Website By ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, website.ID), nil)
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

		var fetched sqlc.Website
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if fetched.ID != website.ID {
			t.Errorf("Expected website ID %d, got %d", website.ID, fetched.ID)
		}
		if fetched.Name != website.Name {
			t.Errorf("Expected website name %q, got %q", website.Name, fetched.Name)
		}
		if fetched.Domain != website.Domain {
			t.Errorf("Expected website domain %q, got %q", website.Domain, fetched.Domain)
		}
	})
}

func adminUpdateWebsite(t *testing.T, website sqlc.Website, addr string, token string) {
	t.Run("Update Website", func(t *testing.T) {
		updateReq := testUpdateWebsiteParams(website.ID)

		body, err := json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", addr, website.ID), bytes.NewBuffer(body))
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

		var updated sqlc.Website
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if updated.Name != updateReq.Name {
			t.Errorf("Expected name %q, got %q", updateReq.Name, updated.Name)
		}
		if updated.Domain != updateReq.Domain {
			t.Errorf("Expected domain %q, got %q", updateReq.Domain, updated.Domain)
		}
		if updated.Status != updateReq.Status {
			t.Errorf("Expected status %q, got %q", updateReq.Status, updated.Status)
		}
	})
}

func adminDeleteWebsite(t *testing.T, website sqlc.Website, addr string, token string) {
	t.Run("Delete Website", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", addr, website.ID), nil)
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

func adminVerifyWebsiteDeleted(t *testing.T, website sqlc.Website, addr string, token string) {
	t.Run("Verify Website Deleted", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, website.ID), nil)
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
