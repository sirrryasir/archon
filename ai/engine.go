package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirrryasir/archon/config"
	"github.com/spf13/viper"
)

// StreamHandler is a callback that receives text chunks as they are generated.
type StreamHandler func(chunk string)

// Engine defines the interface for interacting with various LLM providers.
type Engine interface {
	ChatStream(ctx context.Context, messages []Message, onChunk StreamHandler) (string, error)
}

// FallbackProvider wraps multiple engines for high availability.
type FallbackProvider struct {
	primary  Engine
	fallback Engine
}

func (p *FallbackProvider) ChatStream(ctx context.Context, messages []Message, onChunk StreamHandler) (string, error) {
	// Try primary
	resp, err := p.primary.ChatStream(ctx, messages, onChunk)
	if err != nil {
		if p.fallback != nil && !strings.Contains(err.Error(), "terminated by guardian") {
			onChunk("\n\n[Primary provider failed. Attempting fallback...]\n\n")
			return p.fallback.ChatStream(ctx, messages, onChunk)
		}
		return resp, err
	}
	return resp, nil
}

// NewEngine initializes the correct provider based on configuration.
func NewEngine() Engine {
	providerName := config.GetProvider()
	modelName := config.GetModel()

	primary := getEngineFor(providerName, modelName)

	fallbackModel := viper.GetString("fallback_model")
	fallbackProvider := viper.GetString("fallback_provider")

	if fallbackModel != "" && fallbackProvider != "" {
		return &FallbackProvider{
			primary:  primary,
			fallback: getEngineFor(fallbackProvider, fallbackModel),
		}
	}

	return primary
}

func getEngineFor(provider, model string) Engine {
	// Check for aliases first
	alias := ResolveModelAlias(model)
	if alias != nil {
		provider = alias.Provider
		model = alias.Model
	}

	switch provider {
	case "anthropic":
		p := NewAnthropicProvider()
		p.model = model
		return p
	case "google":
		p := NewGoogleProvider()
		p.model = model
		return p
	case "ollama":
		clientConfig := openai.DefaultConfig("ollama-dummy-key")
		clientConfig.BaseURL = fmt.Sprintf("%s/v1", config.GetOllamaURL())
		return &OpenAIProvider{client: openai.NewClientWithConfig(clientConfig), model: model}
	default:
		// Default to OpenAI
		clientConfig := openai.DefaultConfig(config.GetOpenAIKey())
		return &OpenAIProvider{client: openai.NewClientWithConfig(clientConfig), model: model}
	}
}

// OpenAIProvider implements the Engine interface for OpenAI and Ollama.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// ChatStream executes a streaming chat completion, processing output through the Guardian.
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []Message, onChunk StreamHandler) (string, error) {
	// Convert our domain Messages to OpenAI format
	var oaiMessages []openai.ChatCompletionMessage
	for _, m := range messages {
		oaiMessages = append(oaiMessages, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: oaiMessages,
		Stream:   true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var fullResponse strings.Builder
	var lineBuffer strings.Builder

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fullResponse.String(), err
		}

		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			fullResponse.WriteString(chunk)
			lineBuffer.WriteString(chunk)

			// Simple Guardian Intercept: Check line-by-line
			if strings.Contains(chunk, "\n") {
				line := lineBuffer.String()
				if ScanChunkForViolations(line) {
					// Violation detected! Terminate the stream and inject a warning.
					warning := "\n\n[OUTPUT TERMINATED BY GUARDIAN: Code generation detected.]"
					onChunk(warning)
					return fullResponse.String() + warning, errors.New("output terminated by guardian")
				}
				lineBuffer.Reset()
			}

			// Pass safe chunks to the UI
			onChunk(chunk)
		}
	}

	return fullResponse.String(), nil
}
