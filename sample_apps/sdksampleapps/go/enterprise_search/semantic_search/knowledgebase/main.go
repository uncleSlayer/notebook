package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"pipeshub/models/components"
	"enterprise_search/auth"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <path-to-.env>")
	}
	if err := godotenv.Load(os.Args[1]); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	client, err := auth.NewClient(
		os.Getenv("PIPESHUB_TEST_USER_EMAIL"),
		os.Getenv("PIPESHUB_TEST_USER_PASSWORD"),
	)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.SemanticSearch.Search(context.Background(), components.SemanticSearchRequest{
		Query:   "Who is the business brain?",
		Filters: &components.Filters{Kb: []string{os.Getenv("KB_ID")}},
	})
	if err != nil {
		log.Fatalf("search: %v", err)
	}

	for i, searchResult := range res.SemanticSearchExecuteResponse.SearchResponse.SearchResults {
		name, _ := searchResult.Metadata.RecordName.GetOrZero()
		id, _ := searchResult.Metadata.RecordID.GetOrZero()
		chunk, _ := searchResult.Content.GetOrZero()
		fmt.Printf("─── Result %d ──────────────────────────────────────────────\n", i+1)
		fmt.Printf("  Record:  %s\n", name)
		fmt.Printf("  ID:      %s\n", id)
		fmt.Printf("  Chunk:   %s\n\n", chunk)
	}
}
