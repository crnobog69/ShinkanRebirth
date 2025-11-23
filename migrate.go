package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type OldFormat struct {
	Mangas []map[string]interface{} `json:"mangas"`
}

type NewFormat struct {
	Feeds []map[string]interface{} `json:"feeds"`
}

func main() {
	// Read old format
	data, err := os.ReadFile("./data/mangas.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var oldData OldFormat
	if err := json.Unmarshal(data, &oldData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Convert to new format
	newData := NewFormat{Feeds: make([]map[string]interface{}, 0, len(oldData.Mangas))}

	for _, manga := range oldData.Mangas {
		feed := make(map[string]interface{})
		
		// Copy all fields
		for k, v := range manga {
			feed[k] = v
		}
		
		// Add type field if not present (default to manga)
		if _, ok := feed["type"]; !ok {
			feed["type"] = "manga"
		}
		
		// Add category if not present
		if _, ok := feed["category"]; !ok {
			feed["category"] = "Uncategorized"
		}
		
		// Add addedAt if not present
		if _, ok := feed["addedAt"]; !ok {
			feed["addedAt"] = ""
		}
		
		newData.Feeds = append(newData.Feeds, feed)
	}

	// Write new format
	jsonData, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Backup old file
	if err := os.Rename("./data/mangas.json", "./data/mangas.json.backup"); err != nil {
		fmt.Printf("Error creating backup: %v\n", err)
		os.Exit(1)
	}

	// Write new file
	if err := os.WriteFile("./data/feeds.json", jsonData, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully migrated %d manga(s) to new format\n", len(newData.Feeds))
	fmt.Println("📁 Old data backed up to: ./data/mangas.json.backup")
	fmt.Println("📁 New data written to: ./data/feeds.json")
	fmt.Println("\n💡 Update your .env file:")
	fmt.Println("   DATA_FILE=./data/feeds.json")
}
