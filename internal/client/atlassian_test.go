package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlassian-org-tool/internal/stats"
	"atlassian-org-tool/internal/types"
)

func TestNewAtlassianClient(t *testing.T) {
	config := types.Config{
		AdminBaseURL: "https://api.atlassian.com",
		AccessToken:  "test-token",
		Username:     "test@example.com",
		APIToken:     "test-api-token",
		OrgID:        "test-org-id",
	}

	client := NewAtlassianClient(config)

	if client == nil {
		t.Fatal("NewAtlassianClient returned nil")
	}

	if client.config.AdminBaseURL != config.AdminBaseURL {
		t.Errorf("Expected AdminBaseURL %s, got %s", config.AdminBaseURL, client.config.AdminBaseURL)
	}

	if client.config.AccessToken != config.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", config.AccessToken, client.config.AccessToken)
	}

	if client.httpClient == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestMakeRequestAuthentication(t *testing.T) {
	// Test OAuth authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-oauth-token" {
			t.Errorf("Expected OAuth Authorization header, got: %s", auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	config := types.Config{
		AccessToken: "test-oauth-token",
	}

	client := NewAtlassianClient(config)
	tracker := stats.NewTracker()

	_, err := client.makeRequest("GET", server.URL, nil, true, tracker)
	if err != nil {
		t.Errorf("OAuth request failed: %v", err)
	}

	// Test Basic Auth authentication
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected basic auth headers")
		}
		if username != "test-user" || password != "test-token" {
			t.Errorf("Expected basic auth test-user:test-token, got %s:%s", username, password)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server2.Close()

	config2 := types.Config{
		Username: "test-user",
		APIToken: "test-token",
	}

	client2 := NewAtlassianClient(config2)
	tracker2 := stats.NewTracker()

	_, err = client2.makeRequest("GET", server2.URL, nil, false, tracker2)
	if err != nil {
		t.Errorf("Basic auth request failed: %v", err)
	}
}

func TestMakeRequestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	config := types.Config{
		AccessToken: "test-token",
	}

	client := NewAtlassianClient(config)
	tracker := stats.NewTracker()

	_, err := client.makeRequest("GET", server.URL, nil, true, tracker)
	if err != nil {
		t.Errorf("Request with retries failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Check that stats were updated correctly
	stats := tracker.GetStats()
	if stats.APICallsTotal != 3 {
		t.Errorf("Expected 3 total API calls, got %d", stats.APICallsTotal)
	}
	if stats.APICallsSucceeded != 1 {
		t.Errorf("Expected 1 succeeded API call, got %d", stats.APICallsSucceeded)
	}
	if stats.RetryAttempts != 2 {
		t.Errorf("Expected 2 retry attempts, got %d", stats.RetryAttempts)
	}
}

func TestMakeRequestClientErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	config := types.Config{
		AccessToken: "invalid-token",
	}

	client := NewAtlassianClient(config)
	tracker := stats.NewTracker()

	_, err := client.makeRequest("GET", server.URL, nil, true, tracker)
	if err == nil {
		t.Error("Expected error for 401 response")
	}

	// Should not retry on 4xx errors
	stats := tracker.GetStats()
	if stats.APICallsTotal != 1 {
		t.Errorf("Expected 1 total API call (no retries for 4xx), got %d", stats.APICallsTotal)
	}
	if stats.APICallsFailed != 1 {
		t.Errorf("Expected 1 failed API call, got %d", stats.APICallsFailed)
	}
}

func TestGetOrganization(t *testing.T) {
	expectedOrg := `{
		"id": "test-org-id",
		"name": "Test Organization",
		"displayName": "Test Org Display",
		"type": "organization"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the URL path
		expectedPath := "/admin/v1/orgs/test-org-id"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedOrg))
	}))
	defer server.Close()

	config := types.Config{
		AdminBaseURL: server.URL,
		AccessToken:  "test-token",
		OrgID:        "test-org-id",
	}

	client := NewAtlassianClient(config)
	tracker := stats.NewTracker()

	org, err := client.GetOrganization(tracker)
	if err != nil {
		t.Fatalf("GetOrganization failed: %v", err)
	}

	if org.ID != "test-org-id" {
		t.Errorf("Expected org ID 'test-org-id', got '%s'", org.ID)
	}
	if org.Name != "Test Organization" {
		t.Errorf("Expected org name 'Test Organization', got '%s'", org.Name)
	}

	// Check that stats were updated
	stats := tracker.GetStats()
	if stats.OrganizationsFound != 1 {
		t.Errorf("Expected 1 organization found, got %d", stats.OrganizationsFound)
	}
}
