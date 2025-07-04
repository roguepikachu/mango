package main

import (
	"github.com/spf13/cobra"

	"github.com/example/mango/internal/llmselector"
	"github.com/example/mango/internal/orchestrator"
)

var (
	diffRange string
	mode      string
	llmToken  string
	provider  string
)

var rootCmd = &cobra.Command{
	Use:   "mango",
	Short: "Smart test runner",
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&diffRange, "diff", "HEAD~1", "git diff range")
	rootCmd.PersistentFlags().StringVar(&mode, "mode", "auto", "execution mode: auto, go, ginkgo")
	rootCmd.PersistentFlags().StringVar(&llmToken, "llm-token", "", "LLM API token")
	rootCmd.PersistentFlags().StringVar(&provider, "provider", string(llmselector.ProviderOpenAI), "LLM provider: openai, anthropic, gemini")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dryRunCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run selected tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		sel := llmselector.NewSelector(llmselector.Provider(provider), llmToken)
		orch := orchestrator.Orchestrator{Selector: sel, Mode: mode}
		return orch.Run(cmd.Context(), diffRange)
	},
}

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Preview selected tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		sel := llmselector.NewSelector(llmselector.Provider(provider), llmToken)
		orch := orchestrator.Orchestrator{Selector: sel, Mode: mode, DryRun: true}
		return orch.Run(cmd.Context(), diffRange)
	},
}
