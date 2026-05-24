package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Archon configuration for this project",
	Long: `Create a local .archon/config.json configuration file for this project.
This file allows you to set a different provider and model per project,
overriding your global ~/.archon/config.json settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func runInit() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Archon Init — Setting up your project")
	fmt.Println()

	// Check if already initialized
	cwd, _ := os.Getwd()
	configDir := filepath.Join(cwd, ".archon")
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("  ⚠️  .archon/config.json already exists.\n")
		fmt.Print("  Overwrite? [y/N]: ")
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("  Cancelled.")
			return
		}
	}

	// Choose provider
	fmt.Println("  Choose a provider:")
	fmt.Println("    1. ollama    (local models — no API key needed)")
	fmt.Println("    2. anthropic (Claude — requires API key)")
	fmt.Println("    3. openai    (GPT — requires API key)")
	fmt.Println("    4. google    (Gemini — requires API key)")
	fmt.Print("\n  Provider [1]: ")
	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)

	providerMap := map[string]string{
		"1": "ollama",
		"2": "anthropic",
		"3": "openai",
		"4": "google",
	}

	provider := providerMap[providerChoice]
	if provider == "" {
		provider = "ollama"
	}

	// Choose model
	modelDefaults := map[string]string{
		"ollama":    "qwen2.5:7b",
		"anthropic": "claude-3-5-sonnet-latest",
		"openai":    "gpt-4o",
		"google":    "gemini-1.5-pro-latest",
	}
	defaultModel := modelDefaults[provider]

	fmt.Printf("\n  Model [%s]: ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}

	// Build config
	cfg := map[string]string{
		"provider": provider,
		"model":    model,
	}

	// Create directory and write config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("  ❌ Error creating .archon directory: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Printf("  ❌ Error writing config: %v\n", err)
		os.Exit(1)
	}

	// Add to .gitignore if applicable
	gitignorePath := filepath.Join(cwd, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		content, _ := os.ReadFile(gitignorePath)
		if !strings.Contains(string(content), ".archon/") {
			f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				fmt.Fprintf(f, "\n# Archon local config (may contain API keys)\n.archon/\n")
				f.Close()
				fmt.Printf("  ✅ Added .archon/ to .gitignore\n")
			}
		}
	}

	fmt.Printf("\n  ✅ Created %s\n", configPath)
	fmt.Printf("     Provider : %s\n", provider)
	fmt.Printf("     Model    : %s\n", model)
	fmt.Printf("\n  Next: Run `archon doctor` to verify connectivity.\n")
}
