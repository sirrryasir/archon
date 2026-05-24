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

type GoogleProvider struct {
	model string
}

func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		model: config.GetModel(),
	}
}

func (p *GoogleProvider) ChatStream(ctx context.Context, messages []Message, onChunk StreamHandler) (string, error) {
	apiKey := config.GetGoogleKey()
	if apiKey == "" {
		return "", errors.New("GOOGLE_GENERATIVE_AI_API_KEY is not set")
	}

	var systemPrompt string
	var contents []map[string]interface{}

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt += m.Content + "\n"
		} else {
			role := "user"
			if m.Role == "assistant" {
				role = "model"
			}
			contents = append(contents, map[string]interface{}{
				"role": role,
				"parts": []map[string]interface{}{
					{"text": m.Content},
				},
			})
		}
	}

	reqPayload := map[string]interface{}{
		"contents": contents,
	}

	if systemPrompt != "" {
		reqPayload["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": systemPrompt},
			},
		}
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.model, apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("google api error: status %d, %v", resp.StatusCode, errResp)
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullResponse strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &event); err == nil {
				if candidates, ok := event["candidates"].([]interface{}); ok && len(candidates) > 0 {
					if candidate, ok := candidates[0].(map[string]interface{}); ok {
						if content, ok := candidate["content"].(map[string]interface{}); ok {
							if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
								if part, ok := parts[0].(map[string]interface{}); ok {
									if text, ok := part["text"].(string); ok {
										fullResponse.WriteString(text)

										// Guardian: scan full accumulated buffer on every chunk
										if ScanChunkForViolations(fullResponse.String()) {
											warning := "\n\n> ⚠️ **[GUARDIAN INTERCEPTED]**: I don't write implementation code. Let's focus on the architecture. What problem are you actually trying to solve?"
											onChunk(warning)
											return fullResponse.String() + warning, nil
										}
										onChunk(text)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return fullResponse.String(), scanner.Err()
}
