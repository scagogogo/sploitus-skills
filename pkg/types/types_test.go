package types

import (
	"encoding/json"
	"testing"
)

func TestSearchQueryMarshaling(t *testing.T) {
	// Create a SearchQuery
	query := SearchQuery{
		Type:   "exploits",
		Sort:   "default",
		Query:  "test query",
		Title:  true,
		Offset: 10,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("Failed to marshal SearchQuery: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledQuery SearchQuery
	err = json.Unmarshal(jsonData, &unmarshaledQuery)
	if err != nil {
		t.Fatalf("Failed to unmarshal SearchQuery: %v", err)
	}

	// Verify fields
	if unmarshaledQuery.Type != query.Type {
		t.Errorf("Expected Type %s, got %s", query.Type, unmarshaledQuery.Type)
	}
	if unmarshaledQuery.Sort != query.Sort {
		t.Errorf("Expected Sort %s, got %s", query.Sort, unmarshaledQuery.Sort)
	}
	if unmarshaledQuery.Query != query.Query {
		t.Errorf("Expected Query %s, got %s", query.Query, unmarshaledQuery.Query)
	}
	if unmarshaledQuery.Title != query.Title {
		t.Errorf("Expected Title %t, got %t", query.Title, unmarshaledQuery.Title)
	}
	if unmarshaledQuery.Offset != query.Offset {
		t.Errorf("Expected Offset %d, got %d", query.Offset, unmarshaledQuery.Offset)
	}
}

func TestExploitMarshaling(t *testing.T) {
	// Create an Exploit
	exploit := Exploit{
		Title:     "Test Exploit",
		Score:     9.8,
		Href:      "https://example.com/exploit/123",
		Type:      "exploits",
		Published: "2023-05-01",
		ID:        "CVE-2023-12345",
		Source:    "test-source",
		Language:  "python",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(exploit)
	if err != nil {
		t.Fatalf("Failed to marshal Exploit: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledExploit Exploit
	err = json.Unmarshal(jsonData, &unmarshaledExploit)
	if err != nil {
		t.Fatalf("Failed to unmarshal Exploit: %v", err)
	}

	// Verify fields
	if unmarshaledExploit.Title != exploit.Title {
		t.Errorf("Expected Title %s, got %s", exploit.Title, unmarshaledExploit.Title)
	}
	if unmarshaledExploit.Score != exploit.Score {
		t.Errorf("Expected Score %f, got %f", exploit.Score, unmarshaledExploit.Score)
	}
	if unmarshaledExploit.Href != exploit.Href {
		t.Errorf("Expected Href %s, got %s", exploit.Href, unmarshaledExploit.Href)
	}
	if unmarshaledExploit.Type != exploit.Type {
		t.Errorf("Expected Type %s, got %s", exploit.Type, unmarshaledExploit.Type)
	}
	if unmarshaledExploit.Published != exploit.Published {
		t.Errorf("Expected Published %s, got %s", exploit.Published, unmarshaledExploit.Published)
	}
	if unmarshaledExploit.ID != exploit.ID {
		t.Errorf("Expected ID %s, got %s", exploit.ID, unmarshaledExploit.ID)
	}
	if unmarshaledExploit.Source != exploit.Source {
		t.Errorf("Expected Source %s, got %s", exploit.Source, unmarshaledExploit.Source)
	}
	if unmarshaledExploit.Language != exploit.Language {
		t.Errorf("Expected Language %s, got %s", exploit.Language, unmarshaledExploit.Language)
	}
}

func TestSearchResponseMarshaling(t *testing.T) {
	// Create a SearchResponse
	response := SearchResponse{
		Exploits: []Exploit{
			{
				Title:     "Test Exploit 1",
				Score:     9.8,
				Href:      "https://example.com/exploit/123",
				Type:      "exploits",
				Published: "2023-05-01",
				ID:        "CVE-2023-12345",
				Source:    "test-source-1",
				Language:  "python",
			},
			{
				Title:     "Test Exploit 2",
				Score:     7.5,
				Href:      "https://example.com/exploit/456",
				Type:      "exploits",
				Published: "2023-05-02",
				ID:        "CVE-2023-67890",
				Source:    "test-source-2",
				Language:  "c++",
			},
		},
		ExploitsTotal: 2,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal SearchResponse: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaledResponse SearchResponse
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal SearchResponse: %v", err)
	}

	// Verify fields
	if unmarshaledResponse.ExploitsTotal != response.ExploitsTotal {
		t.Errorf("Expected ExploitsTotal %d, got %d", response.ExploitsTotal, unmarshaledResponse.ExploitsTotal)
	}
	if len(unmarshaledResponse.Exploits) != len(response.Exploits) {
		t.Fatalf("Expected %d exploits, got %d", len(response.Exploits), len(unmarshaledResponse.Exploits))
	}

	// Verify first exploit
	if unmarshaledResponse.Exploits[0].Title != response.Exploits[0].Title {
		t.Errorf("Expected Title %s, got %s", response.Exploits[0].Title, unmarshaledResponse.Exploits[0].Title)
	}
	if unmarshaledResponse.Exploits[0].ID != response.Exploits[0].ID {
		t.Errorf("Expected ID %s, got %s", response.Exploits[0].ID, unmarshaledResponse.Exploits[0].ID)
	}

	// Verify second exploit
	if unmarshaledResponse.Exploits[1].Title != response.Exploits[1].Title {
		t.Errorf("Expected Title %s, got %s", response.Exploits[1].Title, unmarshaledResponse.Exploits[1].Title)
	}
	if unmarshaledResponse.Exploits[1].ID != response.Exploits[1].ID {
		t.Errorf("Expected ID %s, got %s", response.Exploits[1].ID, unmarshaledResponse.Exploits[1].ID)
	}
}
