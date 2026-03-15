package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cyperx84/multiplan/internal/config"
	"github.com/cyperx84/multiplan/internal/planner"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan <task>",
	Short: "Run multi-model planning workflow",
	Args:  cobra.MinimumNArgs(1),
	Run:   runPlanCommand,
}

func init() {
	rootCmd.AddCommand(planCmd)
}

func runPlanCommand(cmd *cobra.Command, args []string) {
	task := strings.Join(args, " ")

	cfg := &config.Config{
		Task:          task,
		Requirements:  cmd.Flag("req").Value.String(),
		Constraints:   cmd.Flag("con").Value.String(),
		OutputDir:     cmd.Flag("out").Value.String(),
		DebateModel:   cmd.Flag("debate-model").Value.String(),
		ConvergeModel: cmd.Flag("converge-model").Value.String(),
		Verbose:       cmd.Flag("verbose").Value.String() == "true",
	}

	// Parse timeout
	timeoutFlag := cmd.Flag("timeout").Value.String()
	fmt.Sscanf(timeoutFlag, "%d", &cfg.TimeoutMs)

	// Parse models
	modelsStr := cmd.Flag("models").Value.String()
	if modelsStr != "" {
		cfg.Models = strings.Split(modelsStr, ",")
	}

	// Run planning
	result, err := planner.Run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("\n════════════════════════════════════════")
	fmt.Println(" Multimodel Planning Complete")
	fmt.Println("════════════════════════════════════════")
	fmt.Printf("\n Task: %s\n", task)
	fmt.Println(" Outputs:")

	for _, plan := range result.Plans {
		if plan.Error == "" {
			fmt.Printf("   %s: %s/plan-%s.md\n", plan.ModelName, result.OutputDir, plan.ModelID)
		} else {
			fmt.Printf("   %s: FAILED — %s\n", plan.ModelName, plan.Error)
		}
	}

	fmt.Printf("   Debate:     %s/debate.md\n", result.OutputDir)
	fmt.Printf("   Final Plan: %s/final-plan.md\n", result.OutputDir)
	fmt.Printf("\n%s\n", result.FinalPlan)
}
