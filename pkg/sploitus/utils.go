package sploitus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

// ExportJSON exports the search results to a JSON file
func ExportJSON(response *types.SearchResponse, outputPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the response with indentation for readability
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// DefaultSearchTypes returns the default search types for Sploitus
func DefaultSearchTypes() []string {
	return []string{"exploits", "tools"}
}

// DefaultOutputPath generates a default output file path based on the query and current time
func DefaultOutputPath(query string, searchType string) string {
	timestamp := time.Now().Format("20060102_150405")
	sanitizedQuery := sanitizeFilename(query)
	return fmt.Sprintf("results/%s_%s_%s.json", sanitizedQuery, searchType, timestamp)
}

// sanitizeFilename removes invalid characters from a filename
func sanitizeFilename(name string) string {
	// Replace spaces and special characters
	replacer := map[rune]rune{
		' ':  '_',
		'/':  '_',
		'\\': '_',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	result := []rune{}
	for _, r := range name {
		if replacement, ok := replacer[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
