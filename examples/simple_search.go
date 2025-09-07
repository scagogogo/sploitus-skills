package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scagogogo/sploitus-crawler/pkg/sploitus"
)

func main() {
	// Check if we have a search query
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run simple_search.go <search_query>")
		os.Exit(1)
	}

	// Get the search query from command line
	query := os.Args[1]

	// Create a new client
	client := sploitus.NewClient()

	// Perform a search
	fmt.Printf("Searching for %q...\n", query)
	response, err := client.Search(query, "exploits", "default", 0)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print results
	fmt.Printf("Found %d results\n", response.ExploitsTotal)
	for i, exploit := range response.Exploits {
		fmt.Printf("\n%d. %s\n", i+1, exploit.Title)
		fmt.Printf("   Score: %.1f\n", exploit.Score)
		fmt.Printf("   Type: %s\n", exploit.Type)
		fmt.Printf("   URL: %s\n", exploit.Href)
		fmt.Printf("   Published: %s\n", exploit.Published)
		fmt.Printf("   ID: %s\n", exploit.ID)
	}

	// Export to JSON
	outputPath := sploitus.DefaultOutputPath(query, "exploits")
	if err := sploitus.ExportJSON(response, outputPath); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}
	fmt.Printf("\nResults saved to %s\n", outputPath)
}
