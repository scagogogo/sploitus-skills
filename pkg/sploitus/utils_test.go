package sploitus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

func TestExportJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sploitus-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data
	response := &types.SearchResponse{
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

	// Test export
	outputPath := filepath.Join(tempDir, "test-output.json")
	err = ExportJSON(response, outputPath)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var loadedResponse types.SearchResponse
	err = json.Unmarshal(data, &loadedResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal output JSON: %v", err)
	}

	// Check content
	if loadedResponse.ExploitsTotal != response.ExploitsTotal {
		t.Errorf("Expected ExploitsTotal %d, got %d", response.ExploitsTotal, loadedResponse.ExploitsTotal)
	}

	if len(loadedResponse.Exploits) != len(response.Exploits) {
		t.Fatalf("Expected %d exploits, got %d", len(response.Exploits), len(loadedResponse.Exploits))
	}

	if loadedResponse.Exploits[0].Title != response.Exploits[0].Title {
		t.Errorf("Expected title %s, got %s", response.Exploits[0].Title, loadedResponse.Exploits[0].Title)
	}
}

func TestExportJSONInvalidPath(t *testing.T) {
	// Test with an invalid path (permission denied or non-existent directory)
	response := &types.SearchResponse{}
	err := ExportJSON(response, "/invalid-path/that-should-not-exist/file.json")

	// We expect this to fail
	if err == nil {
		t.Error("Expected error for invalid path, but got none")
	}
}

func TestDefaultSearchTypes(t *testing.T) {
	types := DefaultSearchTypes()

	// Check that we get expected default search types
	if len(types) != 2 {
		t.Errorf("Expected 2 search types, got %d", len(types))
	}

	// Check for expected values
	expectedTypes := map[string]bool{
		"exploits": true,
		"tools":    true,
	}

	for _, searchType := range types {
		if !expectedTypes[searchType] {
			t.Errorf("Unexpected search type: %s", searchType)
		}
	}
}

func TestDefaultOutputPath(t *testing.T) {
	query := "test query"
	searchType := "exploits"

	outputPath := DefaultOutputPath(query, searchType)

	// Check that the path is properly formatted
	if !strings.Contains(outputPath, "results/") {
		t.Errorf("Path doesn't contain 'results/' directory: %s", outputPath)
	}

	if !strings.Contains(outputPath, "test_query") {
		t.Errorf("Path doesn't contain sanitized query: %s", outputPath)
	}

	if !strings.Contains(outputPath, searchType) {
		t.Errorf("Path doesn't contain search type: %s", outputPath)
	}

	// Check that timestamp is included
	currentYear := time.Now().Format("2006")
	if !strings.Contains(outputPath, currentYear) {
		t.Errorf("Path doesn't contain current year timestamp: %s", outputPath)
	}
}

func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with space", "with_space"},
		{"file/path:with*special?chars", "file_path_with_special_chars"},
		{"<>|\"", "____"},
		{"normal.txt", "normal.txt"},
	}

	for _, tc := range testCases {
		result := sanitizeFilename(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
