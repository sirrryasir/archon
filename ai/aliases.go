package ai

import "strings"

// ModelAlias represents a resolved provider and model identifier.
type ModelAlias struct {
	Provider string
	Model    string
}

// ResolveModelAlias checks if a string is a known alias and returns the full provider and model.
func ResolveModelAlias(alias string) *ModelAlias {
	alias = strings.ToLower(strings.TrimSpace(alias))

	aliases := map[string]ModelAlias{
		"sonnet": {Provider: "anthropic", Model: "claude-3-5-sonnet-latest"},
		"opus":   {Provider: "anthropic", Model: "claude-3-opus-latest"},
		"haiku":  {Provider: "anthropic", Model: "claude-3-5-haiku-latest"},
		"gpt4o":  {Provider: "openai", Model: "gpt-4o"},
		"flash":  {Provider: "google", Model: "gemini-1.5-flash-latest"},
		"gemini": {Provider: "google", Model: "gemini-1.5-pro-latest"},
		"qwen":   {Provider: "ollama", Model: "qwen2.5-coder:32b"},
		"llama":  {Provider: "ollama", Model: "llama3.1:70b"},
	}

	if resolved, ok := aliases[alias]; ok {
		return &resolved
	}

	return nil
}
