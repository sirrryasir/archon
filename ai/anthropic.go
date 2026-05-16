package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirrryasir/archon/config"
)

type AnthropicProvider struct {
	model string
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		model: config.GetModel(),
	}
}

func (p *AnthropicProvider) ChatStream(ctx context.Context, messages []Message, onChunk StreamHandler) (string, error) {
	apiKey := config.GetAnthropicKey()
	if apiKey == "" {
		return "", errors.New("ANTHROPIC_API_KEY is not set")
	}

	var systemPrompt string
	var chatMessages []map[string]interface{}

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt += m.Content + "\n"
		} else {
			chatMessages = append(chatMessages, map[string]interface{}{
				"role":    m.Role,
				"content": m.Content,
			})
		}
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":      p.model,
		"max_tokens": 8192,
		"system":     systemPrompt,
		"messages":   chatMessages,
		"stream":     true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("anthropic api error: status %d, %v", resp.StatusCode, errResp)
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder
	var lineBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &event); err == nil {
				// Anthropic SSE structure for text chunks
				if t, ok := event["type"].(string); ok && t == "content_block_delta" {
					if delta, ok := event["delta"].(map[string]interface{}); ok {
						if text, ok := delta["text"].(string); ok {
							fullResponse.WriteString(text)
							lineBuffer.WriteString(text)

							// Guardian Intercept
							if strings.Contains(text, "\n") {
								bufStr := lineBuffer.String()
								if ScanChunkForViolations(bufStr) {
									warning := "\n\n[OUTPUT TERMINATED BY GUARDIAN: Code generation detected.]"
									onChunk(warning)
									return fullResponse.String() + warning, errors.New("output terminated by guardian")
								}
								lineBuffer.Reset()
							}
							onChunk(text)
						}
					}
				}
			}
		}
	}

	return fullResponse.String(), scanner.Err()
}
