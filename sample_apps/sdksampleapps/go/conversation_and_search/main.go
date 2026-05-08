// conversation_and_search is an end-to-end demo of the PipesHub Go SDK. It
// authenticates once, then runs a series of scenarios across two domain
// packages:
//
//   - conversation_and_search/conversation — streaming conversations, agent
//     conversations, full lifecycle (archive / unarchive / delete), and
//     KB-filtered variants.
//   - conversation_and_search/search — semantic search, with and without a KB filter.
//
// Each scenario is independent: a failure in one prints a red message but
// does not abort the demo, so you always see results from every section.
//
// Configuration comes from env vars (see .env.example).
package main

import (
	"fmt"
	"os"

	"conversation_and_search/conversation"
	"conversation_and_search/internal/ui"
	"conversation_and_search/pipeshub"
	"conversation_and_search/search"
)

func main() {
	baseURL := os.Getenv("PIPESHUB_BASE_URL")
	email := os.Getenv("PIPESHUB_EMAIL")
	password := os.Getenv("PIPESHUB_PASSWORD")
	query := os.Getenv("PIPESHUB_QUERY")
	agentKey := os.Getenv("AGENT_KEY")
	kbID := os.Getenv("PIPESHUB_KB_ID")
	followUp := "can you give me more details about that?"

	if baseURL == "" || email == "" || password == "" || query == "" {
		ui.Errorf("missing required env vars (PIPESHUB_BASE_URL, PIPESHUB_EMAIL, PIPESHUB_PASSWORD, PIPESHUB_QUERY)")
		os.Exit(1)
	}

	client := pipeshub.NewClient(baseURL)

	// ── Authenticate ──────────────────────────────────────────────────────
	ui.Section("Authenticate")
	user, err := client.Authenticate(email, password)
	if err != nil {
		ui.Errorf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("%s  ✔ Hello, %s%s%s! (email: %s, accountType: %s)\n",
		ui.Green, ui.Bold, user.FullName, ui.Reset, user.Email, user.AccountType)

	// ── User conversation lifecycle ───────────────────────────────────────
	if err := conversation.RunUserLifecycle(client, query, followUp); err != nil {
		ui.Errorf("user lifecycle: %v", err)
	}

	// ── Agent conversation lifecycle ──────────────────────────────────────
	if agentKey == "" {
		ui.Warn("AGENT_KEY not set — skipping agent conversation scenarios")
	} else {
		if err := conversation.RunAgentLifecycle(client, agentKey, query, followUp); err != nil {
			ui.Errorf("agent lifecycle: %v", err)
		}
	}

	// ── KB-filtered scenarios ─────────────────────────────────────────────
	if kbID == "" {
		ui.Warn("PIPESHUB_KB_ID not set — skipping KB-filtered scenarios")
	} else {
		if err := conversation.RunUserKBFiltered(client, query, followUp, kbID); err != nil {
			ui.Errorf("user KB-filtered: %v", err)
		}
		if agentKey != "" {
			if err := conversation.RunAgentKBFiltered(client, agentKey, query, followUp, kbID); err != nil {
				ui.Errorf("agent KB-filtered: %v", err)
			}
		}
		if err := search.RunSemanticSearchWithKBFilter(client, query, 5, kbID); err != nil {
			ui.Errorf("KB-filtered search: %v", err)
		}
	}

	// ── Unscoped semantic search ──────────────────────────────────────────
	if err := search.RunSemanticSearch(client, query, 5); err != nil {
		ui.Errorf("semantic search: %v", err)
	}

	fmt.Println()
}
