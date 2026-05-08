// Package conversation demonstrates the PipesHub SDK's conversation
// surface: streaming chat, follow-up messages, archive/unarchive/delete
// lifecycle, and KB-filtered variants — for both user conversations and
// agent conversations.
//
// Each scenario takes a *pipeshub.Client and runs end-to-end against a live
// backend. Errors are returned, not logged, so the caller decides whether
// to abort the whole demo or move on.
package conversation

import (
	"fmt"

	"conversation_and_search/internal/ui"
	"conversation_and_search/pipeshub"
)

// RunUserLifecycle demonstrates the full user-conversation lifecycle:
// stream a new conversation, send a follow-up, then archive → unarchive →
// delete with count checks at each step.
func RunUserLifecycle(client *pipeshub.Client, query, followUp string) error {
	// New conversation.
	ui.Section("New Conversation")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, query, ui.Reset)

	stream, err := client.StreamConversation(query)
	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	defer stream.Close()
	conversationID := ui.PrintSSE(stream)
	if conversationID == "" {
		return fmt.Errorf("no conversationId returned from stream")
	}

	// Follow-up.
	ui.Section("Add Message")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, followUp, ui.Reset)

	stream2, err := client.AddMessageStream(conversationID, followUp)
	if err != nil {
		return fmt.Errorf("addMessageStream: %w", err)
	}
	defer stream2.Close()
	ui.PrintSSE(stream2)

	// Count → archive → count.
	ui.Section("Conversations — Before Archive")
	before, err := client.GetConversations()
	if err != nil {
		return fmt.Errorf("getConversations: %w", err)
	}
	ui.Info("total conversations: %d", before)

	ui.Section("Archive Conversation")
	archiveResult, err := client.ArchiveConversation(conversationID)
	if err != nil {
		return fmt.Errorf("archiveConversation: %w", err)
	}
	ui.PrintJSON(archiveResult)

	ui.Section("Conversations — After Archive")
	after, err := client.GetConversations()
	if err != nil {
		return fmt.Errorf("getConversations: %w", err)
	}
	ui.Info("total conversations: %d", after)

	// Unarchive.
	ui.Section("Unarchive Conversation")
	unarchiveResult, err := client.UnarchiveConversation(conversationID)
	if err != nil {
		return fmt.Errorf("unarchiveConversation: %w", err)
	}
	ui.PrintJSON(unarchiveResult)

	// Delete + verify.
	ui.Section("Conversations — Before Delete")
	beforeDelete, err := client.GetConversations()
	if err != nil {
		return fmt.Errorf("getConversations: %w", err)
	}
	ui.Info("total conversations: %d", beforeDelete)

	ui.Section("Delete Conversation")
	deleteResult, err := client.DeleteConversation(conversationID)
	if err != nil {
		return fmt.Errorf("deleteConversation: %w", err)
	}
	ui.PrintJSON(deleteResult)

	ui.Section("Conversations — After Delete (Verify)")
	afterDelete, err := client.GetConversations()
	if err != nil {
		return fmt.Errorf("getConversations: %w", err)
	}
	ui.Info("total conversations: %d", afterDelete)
	if afterDelete < beforeDelete {
		ui.Success("conversation successfully deleted (count dropped from %d → %d)", beforeDelete, afterDelete)
	} else {
		ui.Errorf("count unchanged — deletion may not have taken effect")
	}
	return nil
}

// RunAgentLifecycle is the agent-conversation equivalent of
// RunUserLifecycle. Hits /agents/{agentKey}/conversations/* endpoints.
func RunAgentLifecycle(client *pipeshub.Client, agentKey, query, followUp string) error {
	ui.Section("Agent — New Conversation")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, query, ui.Reset)

	stream, err := client.StreamAgentConversation(agentKey, query)
	if err != nil {
		return fmt.Errorf("streamAgentConversation: %w", err)
	}
	defer stream.Close()
	conversationID := ui.PrintSSE(stream)
	if conversationID == "" {
		return fmt.Errorf("no conversationId returned from agent stream")
	}

	ui.Section("Agent — Follow-up Message")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, followUp, ui.Reset)

	stream2, err := client.AddAgentMessageStream(agentKey, conversationID, followUp)
	if err != nil {
		return fmt.Errorf("addAgentMessageStream: %w", err)
	}
	defer stream2.Close()
	ui.PrintSSE(stream2)

	ui.Section("Agent Conversations — Before Archive")
	before, err := client.GetAgentConversations(agentKey)
	if err != nil {
		return fmt.Errorf("getAgentConversations: %w", err)
	}
	ui.Info("total agent conversations: %d", before)

	ui.Section("Archive Agent Conversation")
	archiveResult, err := client.ArchiveAgentConversation(agentKey, conversationID)
	if err != nil {
		return fmt.Errorf("archiveAgentConversation: %w", err)
	}
	ui.PrintJSON(archiveResult)

	ui.Section("Agent Conversations — After Archive")
	after, err := client.GetAgentConversations(agentKey)
	if err != nil {
		return fmt.Errorf("getAgentConversations: %w", err)
	}
	ui.Info("total agent conversations: %d", after)

	ui.Section("Unarchive Agent Conversation")
	unarchiveResult, err := client.UnarchiveAgentConversation(agentKey, conversationID)
	if err != nil {
		return fmt.Errorf("unarchiveAgentConversation: %w", err)
	}
	ui.PrintJSON(unarchiveResult)

	ui.Section("Agent Conversations — Before Delete")
	beforeDelete, err := client.GetAgentConversations(agentKey)
	if err != nil {
		return fmt.Errorf("getAgentConversations: %w", err)
	}
	ui.Info("total agent conversations: %d", beforeDelete)

	ui.Section("Delete Agent Conversation")
	deleteResult, err := client.DeleteAgentConversation(agentKey, conversationID)
	if err != nil {
		return fmt.Errorf("deleteAgentConversation: %w", err)
	}
	ui.PrintJSON(deleteResult)

	ui.Section("Agent Conversations — After Delete (Verify)")
	afterDelete, err := client.GetAgentConversations(agentKey)
	if err != nil {
		return fmt.Errorf("getAgentConversations: %w", err)
	}
	ui.Info("total agent conversations: %d", afterDelete)
	if afterDelete < beforeDelete {
		ui.Success("agent conversation successfully deleted (count dropped from %d → %d)", beforeDelete, afterDelete)
	} else {
		ui.Errorf("count unchanged — deletion may not have taken effect")
	}
	return nil
}

// RunUserKBFiltered scopes a user conversation (initial + follow-up) to a
// single knowledge base via SearchFilters{KB: [kbID]}.
func RunUserKBFiltered(client *pipeshub.Client, query, followUp, kbID string) error {
	filter := &pipeshub.SearchFilters{KB: []string{kbID}}

	ui.Section("Conversation with KB Filter")
	fmt.Printf("  %s%q%s   (kb: %s)\n\n", ui.Bold, query, ui.Reset, kbID)

	stream, err := client.StreamConversationWithFilters(query, filter)
	if err != nil {
		return fmt.Errorf("streamConversationWithFilters: %w", err)
	}
	defer stream.Close()
	conversationID := ui.PrintSSE(stream)
	if conversationID == "" {
		return fmt.Errorf("no conversationId returned from filtered stream")
	}

	ui.Section("Follow-up with KB Filter")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, followUp, ui.Reset)

	stream2, err := client.AddMessageStreamWithFilters(conversationID, followUp, filter)
	if err != nil {
		return fmt.Errorf("addMessageStreamWithFilters: %w", err)
	}
	defer stream2.Close()
	ui.PrintSSE(stream2)
	return nil
}

// RunAgentKBFiltered is the agent-conversation equivalent of
// RunUserKBFiltered.
func RunAgentKBFiltered(client *pipeshub.Client, agentKey, query, followUp, kbID string) error {
	filter := &pipeshub.SearchFilters{KB: []string{kbID}}

	ui.Section("Agent Conversation with KB Filter")
	fmt.Printf("  %s%q%s   (kb: %s)\n\n", ui.Bold, query, ui.Reset, kbID)

	stream, err := client.StreamAgentConversationWithFilters(agentKey, query, filter)
	if err != nil {
		return fmt.Errorf("streamAgentConversationWithFilters: %w", err)
	}
	defer stream.Close()
	conversationID := ui.PrintSSE(stream)
	if conversationID == "" {
		return fmt.Errorf("no conversationId returned from filtered agent stream")
	}

	ui.Section("Agent Follow-up with KB Filter")
	fmt.Printf("  %s%q%s\n\n", ui.Bold, followUp, ui.Reset)

	stream2, err := client.AddAgentMessageStreamWithFilters(agentKey, conversationID, followUp, filter)
	if err != nil {
		return fmt.Errorf("addAgentMessageStreamWithFilters: %w", err)
	}
	defer stream2.Close()
	ui.PrintSSE(stream2)
	return nil
}
