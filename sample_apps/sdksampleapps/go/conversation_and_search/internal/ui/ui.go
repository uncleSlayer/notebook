// Package ui contains shared console-presentation helpers used by the sample
// app: section headers, colored output, JSON pretty-printing, and an SSE
// stream parser. Kept under internal/ so it stays out of the sample's public
// surface — readers should focus on the SDK calls, not console formatting.
package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ANSI color codes.
const (
	Cyan   = "\033[36m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Bold   = "\033[1m"
	Reset  = "\033[0m"
)

// Section prints a visually distinct section header. Includes a small delay
// so the previous section's output has settled before the next one starts.
func Section(title string) {
	time.Sleep(2 * time.Second)
	fmt.Printf("\n%s%s┌─ %s %s\n", Bold, Cyan, title, Reset)
}

// PrintJSON pretty-prints any value as indented JSON in yellow.
func PrintJSON(v any) {
	b, _ := json.MarshalIndent(v, "  ", "  ")
	fmt.Printf("%s  %s%s\n", Yellow, string(b), Reset)
}

// Success prints a green success message.
func Success(format string, args ...any) {
	fmt.Printf("%s  ✔ "+format+"%s\n", append([]any{Green}, append(args, Reset)...)...)
}

// Warn prints a yellow warning message.
func Warn(format string, args ...any) {
	fmt.Printf("%s  ⚠ "+format+"%s\n", append([]any{Yellow}, append(args, Reset)...)...)
}

// Errorf prints a red error message.
func Errorf(format string, args ...any) {
	fmt.Printf("%s  ✖ "+format+"%s\n", append([]any{Red}, append(args, Reset)...)...)
}

// Info prints a yellow info bullet.
func Info(format string, args ...any) {
	fmt.Printf("%s  ● "+format+"%s\n", append([]any{Yellow}, append(args, Reset)...)...)
}

// PrintSSE consumes a Server-Sent Events stream from a PipesHub conversation
// endpoint and prints each event in a human-readable form. It returns the
// conversationId discovered in the "connected" event so callers can issue
// follow-up messages.
func PrintSSE(stream io.ReadCloser) (conversationID string) {
	scanner := bufio.NewScanner(stream)
	var event string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			switch event {
			case "connected":
				var p map[string]any
				json.Unmarshal([]byte(data), &p)
				if id, ok := p["conversationId"].(string); ok {
					conversationID = id
				}
				fmt.Printf("%s  ● connected  conversationId: %s%s\n", Green, conversationID, Reset)

			case "status":
				var p map[string]any
				json.Unmarshal([]byte(data), &p)
				fmt.Printf("%s  ↻ %v%s\n", Yellow, p["message"], Reset)

			case "answer_chunk":
				var p map[string]any
				json.Unmarshal([]byte(data), &p)
				if chunk, ok := p["chunk"].(string); ok && chunk != "" {
					fmt.Print(chunk)
				}

			case "complete":
				var p map[string]any
				json.Unmarshal([]byte(data), &p)
				confidence := ""
				if conv, ok := p["conversation"].(map[string]any); ok {
					msgs, _ := conv["messages"].([]any)
					for _, m := range msgs {
						msg, _ := m.(map[string]any)
						if msg["messageType"] == "bot_response" {
							confidence, _ = msg["confidence"].(string)
						}
					}
				}
				fmt.Printf("\n\n%s  ✔ complete  confidence: %s%s\n", Green, confidence, Reset)

			case "error":
				fmt.Printf("\n%s  ✖ error: %s%s\n", Red, data, Reset)
			}
		}
	}
	return conversationID
}
