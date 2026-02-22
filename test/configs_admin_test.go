package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-wordpress/internal/auth"
	"go-wordpress/internal/config"
	"go-wordpress/internal/configs/dto"
	"go-wordpress/internal/storage/sql/sqlc"
	"io"
	"net/http"
	"testing"
)

func TestConfigsAdmin(t *testing.T) {
	WithHttpTestServer(t, func() {
		cfg, err := config.NewConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		base := fmt.Sprintf("http://%s:%d/api/v1/admin", cfg.HTTPAddress, cfg.HTTPPort)
		configAddr := base + "/configs"

		token, err := auth.GenerateToken(cfg, "admin-123")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Config has FK on website_id only.
		var website sqlc.Website
		adminCreateWebsite(t, &website, base+"/websites", token)
		if website.ID == 0 {
			t.Fatalf("Failed to create website for config tests")
		}
		defer adminDeleteWebsite(t, website, base+"/websites", token)

		var appConfig sqlc.CreateConfigRow

		adminUnAuthorizedCreateConfig(t, configAddr)
		adminUnAuthorizedListConfig(t, configAddr)
		adminCreateConfig(t, &appConfig, configAddr, token, website.ID)
		adminListConfig(t, appConfig, configAddr, token)
		adminGetConfigByID(t, appConfig, configAddr, token)
		adminUpdateConfig(t, appConfig, configAddr, token, website.ID)
		adminDeleteConfig(t, appConfig, configAddr, token)
		adminVerifyConfigDeleted(t, appConfig, configAddr, token)
	})
}

func testCreateConfigParams(websiteID int32) sqlc.CreateConfigParams {
	return sqlc.CreateConfigParams{
		WebsiteID: websiteID,
		Key:       "test_key",
		Value:     []byte(`{"setting": "test_value"}`),
	}
}

func testUpdateConfigParams(id, websiteID int32) sqlc.UpdateConfigParams {
	return sqlc.UpdateConfigParams{
		ID:        id,
		WebsiteID: websiteID,
		Key:       "updated_key",
		Value:     []byte(`{"setting": "updated_value"}`),
		Status:    sqlc.EntityStatusInactive,
	}
}

func adminUnAuthorizedCreateConfig(t *testing.T, addr string) {
	t.Run("Unauthorized Create Config", func(t *testing.T) {
		body, err := json.Marshal(testCreateConfigParams(1))
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
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

func adminUnAuthorizedListConfig(t *testing.T, addr string) {
	t.Run("Unauthorized List Configs", func(t *testing.T) {
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

func adminCreateConfig(t *testing.T, appConfig *sqlc.CreateConfigRow, addr string, token string, websiteID int32) {
	t.Run("Create Config", func(t *testing.T) {
		data := testCreateConfigParams(websiteID)
		body, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
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

		if err := json.NewDecoder(resp.Body).Decode(appConfig); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if appConfig.ID == 0 {
			t.Fatalf("Expected a non-zero config ID after creation")
		}
		if appConfig.Key != testCreateConfigParams(websiteID).Key {
			t.Errorf("Expected key %q, got %q", testCreateConfigParams(websiteID).Key, appConfig.Key)
		}
		if appConfig.WebsiteID != websiteID {
			t.Errorf("Expected website ID %d, got %d", websiteID, appConfig.WebsiteID)
		}
		t.Logf("Created config: %+v", appConfig)
	})
}

func adminListConfig(t *testing.T, appConfig sqlc.CreateConfigRow, addr string, token string) {
	t.Run("List Configs", func(t *testing.T) {
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

		var configs dto.ConfigsResponse
		if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if len(configs) == 0 {
			t.Fatalf("Expected at least one config, got 0")
		}

		for i, c := range configs {
			if c.ID == appConfig.ID {
				t.Logf("Found created config at index %d: %+v", i, c)
				return
			}
		}
		t.Errorf("Created config (ID=%d) not found in list", appConfig.ID)
	})
}

func adminGetConfigByID(t *testing.T, appConfig sqlc.CreateConfigRow, addr string, token string) {
	t.Run("Get Config By ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, appConfig.ID), nil)
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

		var fetched sqlc.GetConfigByIDRow
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if fetched.ID != appConfig.ID {
			t.Errorf("Expected config ID %d, got %d", appConfig.ID, fetched.ID)
		}
		if fetched.Key != appConfig.Key {
			t.Errorf("Expected config key %q, got %q", appConfig.Key, fetched.Key)
		}
		if fetched.WebsiteID != appConfig.WebsiteID {
			t.Errorf("Expected website ID %d, got %d", appConfig.WebsiteID, fetched.WebsiteID)
		}
	})
}

func adminUpdateConfig(t *testing.T, appConfig sqlc.CreateConfigRow, addr string, token string, websiteID int32) {
	t.Run("Update Config", func(t *testing.T) {
		updateReq := testUpdateConfigParams(appConfig.ID, websiteID)

		body, err := json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", addr, appConfig.ID), bytes.NewBuffer(body))
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

		var updated sqlc.UpdateConfigRow
		if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
			t.Fatalf(FailedToDecodeMessage, err)
		}

		if updated.Key != updateReq.Key {
			t.Errorf("Expected key %q, got %q", updateReq.Key, updated.Key)
		}
		if updated.WebsiteID != updateReq.WebsiteID {
			t.Errorf("Expected website ID %d, got %d", updateReq.WebsiteID, updated.WebsiteID)
		}
	})
}

func adminDeleteConfig(t *testing.T, appConfig sqlc.CreateConfigRow, addr string, token string) {
	t.Run("Delete Config", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", addr, appConfig.ID), nil)
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

func adminVerifyConfigDeleted(t *testing.T, appConfig sqlc.CreateConfigRow, addr string, token string) {
	t.Run("Verify Config Deleted", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", addr, appConfig.ID), nil)
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
