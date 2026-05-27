package ai

import (
	"context"
	"fmt"
)

const autoCompactThreshold = 0.6

// NeedsCompaction checks if the current message history exceeds the safety threshold for the model.
func NeedsCompaction(messages []Message, model string) bool {
	estimatedTokens := EstimateMessageTokens(messages)
	contextLimit := GetModelContextLimit(model)

	if contextLimit == 0 {
		return false
	}

	ratio := float64(estimatedTokens) / float64(contextLimit)
	return ratio > autoCompactThreshold
}

// CompactHistory splits the conversation, keeping the system prompt and the latest 30%.
// It summarizes the older 70% into a state snapshot.
func CompactHistory(ctx context.Context, engine Engine, messages []Message) ([]Message, error) {
	if len(messages) <= 3 {
		return messages, nil
	}

	// 1. Separate system prompt from the rest
	var systemMsg Message
	var chatHistory []Message

	for _, m := range messages {
		if m.Role == "system" {
			systemMsg = m
		} else {
			chatHistory = append(chatHistory, m)
		}
	}

	if len(chatHistory) < 4 {
		return messages, nil
	}

	// 2. Find the split point (roughly 70% of the chat history)
	splitIndex := int(float64(len(chatHistory)) * 0.7)
	// Ensure we split at a user boundary
	if chatHistory[splitIndex].Role == "assistant" && splitIndex > 0 {
		splitIndex--
	}

	oldHistory := chatHistory[:splitIndex]
	recentHistory := chatHistory[splitIndex:]

	// 3. Generate summary of old history
	summaryPrompt := "Please summarize the following architectural conversation history. Extract the key decisions, technical constraints, and current state. Format it as a concise state snapshot."
	for _, m := range oldHistory {
		summaryPrompt += fmt.Sprintf("\n%s: %s", m.Role, m.Content)
	}

	summaryReq := []Message{
		{Role: "system", Content: "You are a state summarization AI. Extract facts only."},
		{Role: "user", Content: summaryPrompt},
	}

	// We use the engine to generate the summary without streaming to UI
	var fullSummary string
	_, err := engine.ChatStream(ctx, summaryReq, func(chunk string) {
		fullSummary += chunk
	})
	if err != nil {
		return messages, err
	}

	// 4. Reconstruct compacted history
	var compacted []Message
	if systemMsg.Content != "" {
		compacted = append(compacted, systemMsg)
	}

	snapshotMsg := Message{
		Role:    "assistant",
		Content: fmt.Sprintf("<state_snapshot>\n%s\n</state_snapshot>\n\n*Note: The conversation history before this point has been compacted to save context window.*", fullSummary),
	}

	compacted = append(compacted, snapshotMsg)
	compacted = append(compacted, recentHistory...)

	return compacted, nil
}
