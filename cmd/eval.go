package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cyperx84/multiplan/internal/eval"
	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval <file-or-dir>",
	Short: "Evaluate a plan file or run directory",
	Args:  cobra.ExactArgs(1),
	Run:   runEvalCommand,
}

var (
	fixtureFile string
	judgeModel  string
	jsonOutput  bool
	noJudge     bool
)

func init() {
	rootCmd.AddCommand(evalCmd)
	evalCmd.Flags().StringVar(&fixtureFile, "fixture", "", "Fixture JSON file (task, requirements, expectedTopics, minScore)")
	evalCmd.Flags().StringVar(&judgeModel, "judge", "claude", "Model for LLM judge scorer")
	evalCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	evalCmd.Flags().BoolVar(&noJudge, "no-judge", false, "Skip LLM judge scorer (faster, structural only)")
}

func runEvalCommand(cmd *cobra.Command, args []string) {
	path := args[0]

	// Load fixture or create default eval case
	evalCase := &eval.EvalCase{
		Task: "Unknown task",
	}

	if fixtureFile != "" {
		data, err := os.ReadFile(fixtureFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading fixture: %v\n", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(data, evalCase); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing fixture: %v\n", err)
			os.Exit(1)
		}
	}

	// Check if path is file or directory
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Path not found: %s\n", path)
		os.Exit(1)
	}

	opts := &eval.EvalOptions{
		Judge: judgeModel,
	}
	if noJudge {
		opts.Judge = ""
	}

	if info.IsDir() {
		// Eval run directory
		reports, err := eval.EvalRun(path, evalCase, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(reports, "", "  ")
			fmt.Println(string(data))
		} else {
			for _, report := range reports {
				fmt.Println(report.Markdown)
				fmt.Println("---")
			}

			// Summary table
			fmt.Println("\n## Summary")
			fmt.Println("| Model | Score | Pass |")
			fmt.Println("|-------|-------|------|")
			for model, report := range reports {
				score := report.OverallScore * 10
				pass := "✅"
				if !report.Pass {
					pass = "❌"
				}
				fmt.Printf("| %s | %.1f/10 | %s |\n", model, score, pass)
			}
		}
	} else {
		// Eval single file
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		report, err := eval.EvalPlan(string(content), evalCase, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println(report.Markdown)
		}
	}
}
