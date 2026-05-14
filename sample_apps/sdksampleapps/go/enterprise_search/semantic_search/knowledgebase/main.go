package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"undefined/models/components"
	"undefined/models/operations"

	"enterprise_search/auth"
)

const knowledgeBaseName = "SDK-test"

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

	ctx := context.Background()

	name := knowledgeBaseName
	kbsRes, err := client.KnowledgeBases.ListKnowledgeBases(ctx, operations.ListKnowledgeBasesRequest{Search: &name})
	if err != nil {
		log.Fatalf("list knowledge bases: %v", err)
	}
	var kbID string
	for _, kb := range kbsRes.Object.GetKnowledgeBases() {
		if kb.Name == name && kb.ID != nil {
			kbID = *kb.ID
			break
		}
	}
	if kbID == "" {
		log.Fatalf("knowledge base %q not found", name)
	}

	res, err := client.SemanticSearch.Search(ctx, components.SemanticSearchRequest{
		Query:   "Who moved the cheese?",
		Filters: &components.Filters{Kb: []string{kbID}},
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
