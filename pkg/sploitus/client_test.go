package sploitus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", DefaultBaseURL, client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to not be nil")
	}

	if client.HTTPClient.Timeout != DefaultTimeout {
		t.Errorf("Expected Timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
	}

	if client.UserAgent == "" {
		t.Error("Expected UserAgent to not be empty")
	}
}

func TestNewClientWithProxy(t *testing.T) {
	// Setup a mock proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a mock proxy that just responds with success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Proxy reached"))
	}))
	defer proxyServer.Close()

	// Create a client with the mock proxy
	client, err := NewClientWithProxy(proxyServer.URL)

	// Check if client was created successfully
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to not be nil")
	}

	// Verify client properties
	if client.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", DefaultBaseURL, client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to not be nil")
	}

	if client.HTTPClient.Transport == nil {
		t.Error("Expected Transport to not be nil when proxy is set")
	}

	// Verify proxy is set in the transport
	transport, ok := client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Error("Expected Transport to be of type *http.Transport")
	} else {
		// Call the Proxy function to see if it returns a proxy
		proxyURL, err := transport.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}})
		if err != nil {
			t.Errorf("Proxy function returned an error: %v", err)
		}
		if proxyURL == nil {
			t.Error("Expected proxy URL to not be nil")
		}
	}
}

func TestSetProxy(t *testing.T) {
	// Setup a mock proxy server
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a mock proxy that just responds with success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Proxy reached"))
	}))
	defer proxyServer.Close()

	// Create a client with default settings
	client := NewClient()

	// Set the proxy
	err := client.SetProxy(proxyServer.URL)

	// Check if proxy was set successfully
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify proxy is set in the transport
	transport, ok := client.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Error("Expected Transport to be of type *http.Transport")
	} else {
		// Call the Proxy function to see if it returns a proxy
		proxyURL, err := transport.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}})
		if err != nil {
			t.Errorf("Proxy function returned an error: %v", err)
		}
		if proxyURL == nil {
			t.Error("Expected proxy URL to not be nil")
		}
	}

	// Test with invalid proxy URL
	err = client.SetProxy("http://invalid:url:with:too:many:colons")
	if err == nil {
		t.Error("Expected error for invalid proxy URL, got nil")
	}
}

func TestSearchWithQuery(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		if r.URL.Path != SearchEndpoint {
			t.Errorf("Expected path %s, got %s", SearchEndpoint, r.URL.Path)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept header application/json, got %s", r.Header.Get("Accept"))
		}

		if r.Header.Get("User-Agent") == "" {
			t.Error("Expected User-Agent header to not be empty")
		}

		// Test for successful request
		mockResponse := types.SearchResponse{
			Exploits: []types.Exploit{
				{
					Title:     "Test Exploit",
					Score:     9.8,
					Href:      "https://example.com/exploit/123",
					Type:      "exploits",
					Published: "2023-05-01",
					ID:        "CVE-2023-12345",
					Source:    "test-source",
					Language:  "python",
				},
			},
			ExploitsTotal: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient()
	client.BaseURL = server.URL

	query := &types.SearchQuery{
		Type:   "exploits",
		Sort:   "default",
		Query:  "test",
		Title:  false,
		Offset: 0,
	}

	// Execute test
	resp, err := client.SearchWithQuery(query)

	// Verify results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to not be nil")
	}

	if resp.ExploitsTotal != 1 {
		t.Errorf("Expected ExploitsTotal 1, got %d", resp.ExploitsTotal)
	}

	if len(resp.Exploits) != 1 {
		t.Fatalf("Expected 1 exploit, got %d", len(resp.Exploits))
	}

	if resp.Exploits[0].Title != "Test Exploit" {
		t.Errorf("Expected Title 'Test Exploit', got %s", resp.Exploits[0].Title)
	}

	if resp.Exploits[0].ID != "CVE-2023-12345" {
		t.Errorf("Expected ID 'CVE-2023-12345', got %s", resp.Exploits[0].ID)
	}
}

func TestSearch(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := types.SearchResponse{
			Exploits: []types.Exploit{
				{
					Title:     "Test Exploit",
					Score:     9.8,
					Href:      "https://example.com/exploit/123",
					Type:      "exploits",
					Published: "2023-05-01",
					ID:        "CVE-2023-12345",
					Source:    "test-source",
					Language:  "python",
				},
			},
			ExploitsTotal: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient()
	client.BaseURL = server.URL

	// Execute test
	resp, err := client.Search("test", "exploits", "default", 0)

	// Verify results
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to not be nil")
	}

	if resp.ExploitsTotal != 1 {
		t.Errorf("Expected ExploitsTotal 1, got %d", resp.ExploitsTotal)
	}

	if len(resp.Exploits) != 1 {
		t.Fatalf("Expected 1 exploit, got %d", len(resp.Exploits))
	}
}

func TestErrorHandling(t *testing.T) {
	// Setup test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient()
	client.BaseURL = server.URL

	// Execute test
	_, err := client.Search("test", "exploits", "default", 0)

	// Verify error is handled
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err != nil && err.Error() != "API returned non-200 status code: 500" {
		t.Errorf("Expected error message to contain 'non-200 status code: 500', got: %s", err.Error())
	}
}
