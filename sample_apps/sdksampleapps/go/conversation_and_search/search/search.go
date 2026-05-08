// Package search demonstrates the PipesHub SDK's semantic search surface.
// Two scenarios: an unscoped search across the workspace, and a search
// scoped to a single knowledge base via SearchFilters.
package search

import (
	"fmt"

	"conversation_and_search/internal/ui"
	"conversation_and_search/pipeshub"
)

// RunSemanticSearch performs an unscoped semantic search and pretty-prints
// the top results.
func RunSemanticSearch(client *pipeshub.Client, query string, limit int) error {
	ui.Section("Semantic Search")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, query, ui.Reset)

	resp, err := client.SemanticSearch(query, limit)
	if err != nil {
		return fmt.Errorf("semanticSearch: %w", err)
	}
	printResults(resp)
	return nil
}

// RunSemanticSearchWithKBFilter scopes a semantic search to one knowledge
// base via SearchFilters{KB: [kbID]}.
func RunSemanticSearchWithKBFilter(client *pipeshub.Client, query string, limit int, kbID string) error {
	ui.Section("Semantic Search with KB Filter")
	fmt.Printf("  %s%q%s   (kb: %s)\n\n", ui.Bold, query, ui.Reset, kbID)

	filter := &pipeshub.SearchFilters{KB: []string{kbID}}
	resp, err := client.SemanticSearchWithFilters(query, limit, filter)
	if err != nil {
		return fmt.Errorf("semanticSearchWithFilters: %w", err)
	}
	printResults(resp)
	return nil
}

func printResults(resp *pipeshub.SemanticSearchResponse) {
	ui.Success("searchId: %s  (%d results)", resp.SearchID, len(resp.Results))
	for i, r := range resp.Results {
		fmt.Printf("\n%s  [%d] %s%s\n", ui.Bold, i+1, r.Metadata.RecordName, ui.Reset)
		fmt.Printf("      score: %.4f  |  type: %s  |  connector: %s\n",
			r.Metadata.Score, r.CitationType, r.Metadata.Connector)
		content := r.Content
		if len(content) > 200 {
			content = content[:200] + "…"
		}
		fmt.Printf("      %s\n", content)
	}
}
