package pipeshub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL     string
	AccessToken string
	httpClient  *http.Client
}

type UserInfo struct {
	UserID      string
	OrgID       string
	Email       string
	FullName    string
	AccountType string
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) Authenticate(email, password string) (*UserInfo, error) {
	// Step 1: initAuth
	body1, _ := json.Marshal(map[string]string{"email": email})
	resp1, err := c.httpClient.Post(c.BaseURL+"/userAccount/initAuth", "application/json", bytes.NewBuffer(body1))
	if err != nil {
		return nil, fmt.Errorf("initAuth: %w", err)
	}
	io.ReadAll(resp1.Body)
	resp1.Body.Close()

	sessionToken := resp1.Header.Get("x-session-token")
	if sessionToken == "" {
		return nil, fmt.Errorf("initAuth: no session token in response")
	}

	// Step 2: authenticate
	body2, _ := json.Marshal(map[string]any{
		"method":      "password",
		"credentials": map[string]string{"password": password},
	})
	req2, _ := http.NewRequest("POST", c.BaseURL+"/userAccount/authenticate", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("x-session-token", sessionToken)

	resp2, err := c.httpClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	defer resp2.Body.Close()
	data2, _ := io.ReadAll(resp2.Body)

	if resp2.StatusCode != 200 {
		return nil, fmt.Errorf("authenticate: status %d — %s", resp2.StatusCode, data2)
	}

	var authResp map[string]string
	json.Unmarshal(data2, &authResp)
	c.AccessToken = authResp["accessToken"]

	return decodeUserInfo(c.AccessToken)
}

func (c *Client) StreamConversation(query string) (io.ReadCloser, error) {
	body, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", c.BaseURL+"/conversations/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("stream: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) AddMessageStream(conversationID, query string) (io.ReadCloser, error) {
	body, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", c.BaseURL+"/conversations/"+conversationID+"/messages/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("addMessageStream: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("addMessageStream: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) StreamAgentConversation(agentKey, query string) (io.ReadCloser, error) {
	body, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("streamAgentConversation: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("streamAgentConversation: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) AddAgentMessageStream(agentKey, conversationID, query string) (io.ReadCloser, error) {
	body, _ := json.Marshal(map[string]string{"query": query})
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/"+conversationID+"/messages/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("addAgentMessageStream: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("addAgentMessageStream: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func totalCount(result map[string]any) int {
	pagination, _ := result["pagination"].(map[string]any)
	if pagination == nil {
		return 0
	}
	count, _ := pagination["totalCount"].(float64)
	return int(count)
}

func (c *Client) GetConversations() (int, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("getConversations: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("getConversations: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return totalCount(result), nil
}

func (c *Client) GetAgentConversations(agentKey string) (int, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/agents/"+agentKey+"/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("getAgentConversations: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("getAgentConversations: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return totalCount(result), nil
}

func (c *Client) ArchiveAgentConversation(agentKey, conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/"+conversationID+"/archive", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("archiveAgentConversation: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("archiveAgentConversation: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

func (c *Client) UnarchiveAgentConversation(agentKey, conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/"+conversationID+"/unarchive", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unarchiveAgentConversation: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unarchiveAgentConversation: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

func (c *Client) DeleteAgentConversation(agentKey, conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("DELETE", c.BaseURL+"/agents/"+agentKey+"/conversations/"+conversationID, nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deleteAgentConversation: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("deleteAgentConversation: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

func (c *Client) ArchiveConversation(conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("PATCH", c.BaseURL+"/conversations/"+conversationID+"/archive", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("archive: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("archive: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

func (c *Client) UnarchiveConversation(conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("PATCH", c.BaseURL+"/conversations/"+conversationID+"/unarchive", nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unarchive: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unarchive: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

func (c *Client) DeleteConversation(conversationID string) (map[string]any, error) {
	req, _ := http.NewRequest("DELETE", c.BaseURL+"/conversations/"+conversationID, nil)
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deleteConversation: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("deleteConversation: status %d — %s", resp.StatusCode, data)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	return result, nil
}

// ── Filtered Conversations (KB filter) ────────────────────────────────────

type ConversationRequest struct {
	Query         string         `json:"query"`
	SearchFilters *SearchFilters `json:"searchFilters,omitempty"`
}

func (c *Client) StreamConversationWithFilters(query string, filters *SearchFilters) (io.ReadCloser, error) {
	body, _ := json.Marshal(ConversationRequest{Query: query, SearchFilters: filters})
	req, _ := http.NewRequest("POST", c.BaseURL+"/conversations/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("streamConversationWithFilters: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("streamConversationWithFilters: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) AddMessageStreamWithFilters(conversationID, query string, filters *SearchFilters) (io.ReadCloser, error) {
	body, _ := json.Marshal(ConversationRequest{Query: query, SearchFilters: filters})
	req, _ := http.NewRequest("POST", c.BaseURL+"/conversations/"+conversationID+"/messages/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("addMessageStreamWithFilters: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("addMessageStreamWithFilters: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) StreamAgentConversationWithFilters(agentKey, query string, filters *SearchFilters) (io.ReadCloser, error) {
	body, _ := json.Marshal(ConversationRequest{Query: query, SearchFilters: filters})
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("streamAgentConversationWithFilters: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("streamAgentConversationWithFilters: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

func (c *Client) AddAgentMessageStreamWithFilters(agentKey, conversationID, query string, filters *SearchFilters) (io.ReadCloser, error) {
	body, _ := json.Marshal(ConversationRequest{Query: query, SearchFilters: filters})
	req, _ := http.NewRequest("POST", c.BaseURL+"/agents/"+agentKey+"/conversations/"+conversationID+"/messages/stream", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("addAgentMessageStreamWithFilters: %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("addAgentMessageStreamWithFilters: status %d — %s", resp.StatusCode, body)
	}
	return resp.Body, nil
}

// ── Semantic Search ───────────────────────────────────────────────────────

type SearchFilters struct {
	Apps []string `json:"apps,omitempty"`
	KB   []string `json:"kb,omitempty"`
}

type SearchResultMetadata struct {
	RecordName  string   `json:"recordName"`
	RecordID    string   `json:"recordId"`
	MimeType    string   `json:"mimeType"`
	Origin      string   `json:"origin"`
	WebURL      string   `json:"webUrl"`
	Score       float64  `json:"score"`
	Connector   string   `json:"connector"`
	RecordType  string   `json:"recordType"`
	PageNum     []int    `json:"pageNum"`
	BlockNum    []int    `json:"blockNum"`
	Languages   []string `json:"languages"`
	Topics      []string `json:"topics"`
}

type SearchResult struct {
	Content      string               `json:"content"`
	ChunkIndex   int                  `json:"chunkIndex"`
	CitationType string               `json:"citationType"`
	Metadata     SearchResultMetadata `json:"metadata"`
}

type SemanticSearchResponse struct {
	SearchID string
	Results  []SearchResult
}

// SemanticSearch performs a semantic search against POST /search.
// limit ≤ 0 uses the server default (10). Max is 100.
func (c *Client) SemanticSearch(query string, limit int) (*SemanticSearchResponse, error) {
	reqBody := map[string]any{"query": query}
	if limit > 0 {
		reqBody["limit"] = limit
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.BaseURL+"/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("semanticSearch: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("semanticSearch: status %d — %s", resp.StatusCode, data)
	}

	var raw struct {
		SearchID       string `json:"searchId"`
		SearchResponse struct {
			SearchResults []SearchResult `json:"searchResults"`
		} `json:"searchResponse"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("semanticSearch: parse: %w", err)
	}

	return &SemanticSearchResponse{
		SearchID: raw.SearchID,
		Results:  raw.SearchResponse.SearchResults,
	}, nil
}

// SemanticSearchWithFilters performs a semantic search with search filters.
func (c *Client) SemanticSearchWithFilters(query string, limit int, filters *SearchFilters) (*SemanticSearchResponse, error) {
	reqBody := map[string]any{"query": query}
	if limit > 0 {
		reqBody["limit"] = limit
	}
	if filters != nil {
		reqBody["searchFilters"] = filters
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.BaseURL+"/search", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("semanticSearchWithFilters: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("semanticSearchWithFilters: status %d — %s", resp.StatusCode, data)
	}

	var raw struct {
		SearchID       string `json:"searchId"`
		SearchResponse struct {
			SearchResults []SearchResult `json:"searchResults"`
		} `json:"searchResponse"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("semanticSearchWithFilters: parse: %w", err)
	}

	return &SemanticSearchResponse{
		SearchID: raw.SearchID,
		Results:  raw.SearchResponse.SearchResults,
	}, nil
}

func decodeUserInfo(token string) (*UserInfo, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	json.Unmarshal(payload, &claims)

	return &UserInfo{
		UserID:      str(claims["userId"]),
		OrgID:       str(claims["orgId"]),
		Email:       str(claims["email"]),
		FullName:    str(claims["fullName"]),
		AccountType: str(claims["accountType"]),
	}, nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
