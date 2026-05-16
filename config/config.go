package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitConfig sets up the configuration cascade using Viper.
// Resolution order: Defaults > Global File > Local File > Environment > Flags
func InitConfig() {
	// 1. Set Defaults
	viper.SetDefault("provider", "ollama")
	viper.SetDefault("model", "qwen3-coder:480b-cloud")
	viper.SetDefault("ollama_url", "http://localhost:11434")

	// 2. Setup Environment Variables
	viper.SetEnvPrefix("ARCHON")
	viper.AutomaticEnv() // Automatically bind env vars matching ARCHON_*

	// Manually map specific env vars that don't follow the prefix perfectly
	_ = viper.BindEnv("openai_api_key", "OPENAI_API_KEY")
	_ = viper.BindEnv("anthropic_api_key", "ANTHROPIC_API_KEY")
	_ = viper.BindEnv("google_api_key", "GOOGLE_GENERATIVE_AI_API_KEY")
	_ = viper.BindEnv("ollama_url", "OLLAMA_URL")

	// 3. Read Global Config (~/.archon/config.json)
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(home, ".archon"))
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		// We ignore errors here since the global config is optional
		_ = viper.ReadInConfig()
	}

	// 4. Read Local Config (.archon/config.json in current directory)
	cwd, err := os.Getwd()
	if err == nil {
		localViper := viper.New()
		localViper.AddConfigPath(filepath.Join(cwd, ".archon"))
		localViper.SetConfigName("config")
		localViper.SetConfigType("json")
		
		// If local config exists, merge it into the global viper instance
		if err := localViper.ReadInConfig(); err == nil {
			_ = viper.MergeConfigMap(localViper.AllSettings())
		}
	}
}

// GetProvider resolves the current provider based on the configuration cascade.
func GetProvider() string {
	return viper.GetString("provider")
}

// GetModel resolves the current model based on the configuration cascade.
func GetModel() string {
	return viper.GetString("model")
}

// GetOllamaURL returns the Ollama base URL.
func GetOllamaURL() string {
	return viper.GetString("ollama_url")
}

// GetOpenAIKey returns the OpenAI API Key.
func GetOpenAIKey() string {
	return viper.GetString("openai_api_key")
}

// GetAnthropicKey returns the Anthropic API Key.
func GetAnthropicKey() string {
	return viper.GetString("anthropic_api_key")
}

// GetGoogleKey returns the Google API Key.
func GetGoogleKey() string {
	return viper.GetString("google_api_key")
}

// PrintConfig dumps the current resolved configuration for debugging.
func PrintConfig() {
	fmt.Printf("Provider: %s\n", GetProvider())
	fmt.Printf("Model: %s\n", GetModel())
	fmt.Printf("Ollama URL: %s\n", viper.GetString("ollama_url"))
}
