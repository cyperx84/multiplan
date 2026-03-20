package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "multiplan",
	Short: "4-model parallel planning workflow with eval framework",
	Long: `multiplan runs a task through Claude (Opus), Codex (GPT), and GLM-5 simultaneously.
Each produces an independent plan. Then cross-examine them. Then converge on the best synthesis.`,
	Version: "0.5.0",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			runPlanCommand(cmd, args)
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Clean version output: "multiplan v0.5.0"
	rootCmd.SetVersionTemplate("multiplan v{{.Version}}\n")

	rootCmd.PersistentFlags().String("req", "", "Requirements")
	rootCmd.PersistentFlags().String("con", "", "Constraints")
	rootCmd.PersistentFlags().String("out", "", "Output directory")
	rootCmd.PersistentFlags().String("models", "", "Comma-separated models (claude,gemini,codex,glm5)")
	rootCmd.PersistentFlags().String("debate-model", "claude", "Model for debate phase")
	rootCmd.PersistentFlags().String("converge-model", "claude", "Model for convergence phase")
	rootCmd.PersistentFlags().Int("timeout", 300000, "Per-model timeout in milliseconds (default 5m)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all progress output (errors and final result only)")
	rootCmd.PersistentFlags().Bool("skip-lattice", false, "Skip lattice mental model framing (Phase 0)")
	rootCmd.PersistentFlags().String("lattice-cmd", "lattice", "Lattice binary command")
	rootCmd.PersistentFlags().String("claude-model", "", "Model override for Claude CLI mode (e.g. claude-opus-4-20250514)")
	rootCmd.PersistentFlags().String("claude-cmd", "", "Claude CLI binary path (default: claude)")
}
