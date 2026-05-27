package cmd

import (
	"os"

	"github.com/sirrryasir/archon/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is the current release version of Archon. Set via ldflags at build time.
const Version = "0.1.1"

var rootCmd = &cobra.Command{
	Use:     "archon",
	Version: Version,
	Short:   "Archon CLI: The Socratic AI Software Architect",
	Long: `Archon is an elite Socratic AI Software Architect designed to help developers 
build robust, scalable systems through architectural inquiry and guidance.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if m, _ := cmd.Flags().GetString("model"); m != "" {
			viper.Set("model", m)
		}
		if p, _ := cmd.Flags().GetString("provider"); p != "" {
			viper.Set("provider", p)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		StartChatSession()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)

	// Define global flags
	rootCmd.PersistentFlags().StringP("model", "m", "", "AI model override (e.g., sonnet, qwen)")
	rootCmd.PersistentFlags().StringP("provider", "p", "", "AI provider override (e.g., anthropic, ollama)")
}
