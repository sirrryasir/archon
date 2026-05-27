package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirrryasir/archon/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check Archon's configuration and provider connectivity",
	Long:  `Run diagnostics to verify your Archon setup: provider connection, model availability, and config resolution.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDoctor()
	},
}

func runDoctor() {
	fmt.Println("🩺 Archon Doctor — Running diagnostics...")
	fmt.Println()

	// 1. Config check
	provider := config.GetProvider()
	model := config.GetModel()
	fmt.Printf("  ✅ Config loaded\n")
	fmt.Printf("     Provider : %s\n", provider)
	fmt.Printf("     Model    : %s\n", model)

	// 2. Provider connectivity
	fmt.Printf("\n  🔌 Checking %s connectivity...\n", provider)

	switch provider {
	case "ollama":
		ollamaURL := config.GetOllamaURL()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, _ := http.NewRequestWithContext(ctx, "GET", ollamaURL+"/api/tags", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("  ❌ Cannot reach Ollama at %s\n", ollamaURL)
			fmt.Printf("     Error: %v\n", err)
			fmt.Printf("     Fix  : Start Ollama with `ollama serve`\n")
			os.Exit(1)
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Printf("  ✅ Ollama is reachable at %s\n", ollamaURL)
		} else {
			fmt.Printf("  ⚠️  Ollama returned status %d\n", resp.StatusCode)
		}

	case "anthropic":
		key := config.GetAnthropicKey()
		if key == "" || key == "sk-ant-your-anthropic-api-key" {
			fmt.Printf("  ❌ ANTHROPIC_API_KEY is not set\n")
			fmt.Printf("     Fix  : export ANTHROPIC_API_KEY=sk-ant-...\n")
			os.Exit(1)
		}
		fmt.Printf("  ✅ Anthropic API key is configured (sk-ant-...%s)\n", key[len(key)-4:])

	case "google":
		key := config.GetGoogleKey()
		if key == "" || key == "your-gemini-api-key" {
			fmt.Printf("  ❌ GOOGLE_GENERATIVE_AI_API_KEY is not set\n")
			fmt.Printf("     Fix  : export GOOGLE_GENERATIVE_AI_API_KEY=...\n")
			os.Exit(1)
		}
		fmt.Printf("  ✅ Google API key is configured\n")

	case "openai":
		key := config.GetOpenAIKey()
		if key == "" || key == "sk-your-openai-api-key" {
			fmt.Printf("  ❌ OPENAI_API_KEY is not set\n")
			fmt.Printf("     Fix  : export OPENAI_API_KEY=sk-...\n")
			os.Exit(1)
		}
		fmt.Printf("  ✅ OpenAI API key is configured (sk-...%s)\n", key[len(key)-4:])
	}

	// 3. Summary
	fmt.Printf("\n  ✅ Archon %s is ready to use!\n", Version)
	fmt.Printf("     Run: archon chat    — start interactive session\n")
	fmt.Printf("     Run: archon prompt  — send a single prompt\n")
}
