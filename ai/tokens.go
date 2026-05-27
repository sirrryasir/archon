package ai

import (
	"fmt"
	"strings"
)

// Message represents a unified chat message structure across all providers.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// EstimateTokens calculates an approximate token count based on character length.
// The industry standard heuristic is 1 token ~= 4 characters for English text.
func EstimateTokens(text string) int {
	return len(text) / 4
}

// EstimateMessageTokens calculates the approximate total token count for a slice of messages.
func EstimateMessageTokens(messages []Message) int {
	var total int
	for _, msg := range messages {
		total += EstimateTokens(msg.Content)
		// Add some padding for JSON/message structural overhead
		total += 4
	}
	return total
}

// GetModelContextLimit returns the maximum context window size for popular models.
func GetModelContextLimit(model string) int {
	modelLower := strings.ToLower(model)

	// Anthropic
	if strings.Contains(modelLower, "claude-3-5") || strings.Contains(modelLower, "opus") || strings.Contains(modelLower, "sonnet") {
		return 200_000
	}

	// Google
	if strings.Contains(modelLower, "gemini-1.5-pro") {
		return 2_000_000
	}
	if strings.Contains(modelLower, "gemini-1.5-flash") {
		return 1_000_000
	}

	// OpenAI
	if strings.Contains(modelLower, "gpt-4o") || strings.Contains(modelLower, "gpt-4-turbo") {
		return 128_000
	}

	// Default generic fallback (e.g. for standard open-source Ollama models)
	// We use 32k as a safe baseline for modern local models like Qwen/Llama3
	return 32_000
}

// FormatTokenCount formats a token number into a human-readable string (e.g., 12.5k)
func FormatTokenCount(tokens int) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000.0)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%.1fk", float64(tokens)/1000.0)
	}
	return fmt.Sprintf("%d", tokens)
}
