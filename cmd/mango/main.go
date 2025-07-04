package main

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/example/mango/internal/advisor"
	"github.com/example/mango/internal/diff"
	"github.com/example/mango/internal/generator"
	"github.com/example/mango/internal/llmselector"
	"github.com/example/mango/internal/orchestrator"
	"github.com/example/mango/internal/predictor"
	"github.com/example/mango/internal/query"
	"github.com/example/mango/internal/testmeta"
)

var (
	diffRange string
	mode      string
	llmToken  string
	provider  string
	planDesc  string
	question  string
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
	rootCmd.PersistentFlags().StringVar(&planDesc, "plan", "", "planned change description")
	rootCmd.PersistentFlags().StringVar(&question, "question", "", "query question")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(dryRunCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(predictCmd)
	rootCmd.AddCommand(adviceCmd)
	rootCmd.AddCommand(queryCmd)
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

var generateCmd = &cobra.Command{
	Use:   "generate-tests",
	Short: "Generate new test scenarios",
	RunE: func(cmd *cobra.Command, args []string) error {
		sel := generator.New(generator.NewOpenAIClient(llmToken))
		changes, err := diff.AnalyzeDiff(diffRange)
		if err != nil {
			return err
		}
		tests, err := testmeta.Extract()
		if err != nil {
			return err
		}
		names, err := sel.Generate(cmd.Context(), changes, tests)
		if err != nil {
			return err
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	},
}

var predictCmd = &cobra.Command{
	Use:   "predict",
	Short: "Predict failing tests from a plan description",
	RunE: func(cmd *cobra.Command, args []string) error {
		tests, err := testmeta.Extract()
		if err != nil {
			return err
		}
		p := predictor.New(predictor.NewOpenAIClient(llmToken))
		names, err := p.Predict(cmd.Context(), planDesc, tests)
		if err != nil {
			return err
		}
		for _, n := range names {
			fmt.Println(n)
		}
		return nil
	},
}

var adviceCmd = &cobra.Command{
	Use:   "advise",
	Short: "Get code quality advice",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := advisor.New(advisor.NewOpenAIClient(llmToken))
		msg, err := a.Advise(cmd.Context())
		if err != nil {
			return err
		}
		fmt.Println(msg)
		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query tests via natural language",
	RunE: func(cmd *cobra.Command, args []string) error {
		tests, err := testmeta.Extract()
		if err != nil {
			return err
		}
		q := query.New(query.NewOpenAIClient(llmToken))
		msg, err := q.Ask(cmd.Context(), question, tests)
		if err != nil {
			return err
		}
		fmt.Println(msg)
		return nil
	},
}
