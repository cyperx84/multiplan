package cmd

import (
	"encoding/json"
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
	planCmd.Flags().Bool("json", false, "Output structured JSON result")
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
		Quiet:         cmd.Flag("quiet").Value.String() == "true",
		SkipLattice:   cmd.Flag("skip-lattice").Value.String() == "true",
		LatticeCmd:    cmd.Flag("lattice-cmd").Value.String(),
	}

	// --json flag (plan subcommand only)
	jsonFlag := false
	if jf := cmd.Flags().Lookup("json"); jf != nil {
		jsonFlag = jf.Value.String() == "true"
	}
	cfg.JSON = jsonFlag

	// Parse timeout
	timeoutFlag := cmd.Flag("timeout").Value.String()
	fmt.Sscanf(timeoutFlag, "%d", &cfg.TimeoutMs)

	// Parse models
	modelsStr := cmd.Flag("models").Value.String()
	if modelsStr != "" {
		cfg.Models = strings.Split(modelsStr, ",")
	}

	// Load config file and apply (CLI flags take precedence)
	fc, err := config.LoadFileConfig()
	if err == nil {
		config.ApplyFileConfig(cfg, fc)
	}

	// Run planning
	result, err := planner.Run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if jsonFlag {
		outputJSON(result)
		return
	}

	// Human-readable output
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

// planJSONOutput is the structured JSON response for --json flag on plan.
type planJSONOutput struct {
	RunID         string          `json:"run_id"`
	OutputDir     string          `json:"output_dir"`
	Models        []planJSONModel `json:"models"`
	DebateExcerpt string          `json:"debate_excerpt"`
	FinalPlan     string          `json:"final_plan"`
}

type planJSONModel struct {
	ModelID     string `json:"model_id"`
	ModelName   string `json:"model_name"`
	PlanExcerpt string `json:"plan_excerpt"`
	DurationMs  int64  `json:"duration_ms"`
	Error       string `json:"error,omitempty"`
}

func outputJSON(result *planner.RunResult) {
	out := planJSONOutput{
		RunID:         result.RunID,
		OutputDir:     result.OutputDir,
		DebateExcerpt: excerpt(result.Debate, 500),
		FinalPlan:     result.FinalPlan,
	}

	for _, p := range result.Plans {
		out.Models = append(out.Models, planJSONModel{
			ModelID:     p.ModelID,
			ModelName:   p.ModelName,
			PlanExcerpt: excerpt(p.Plan, 500),
			DurationMs:  p.DurationMs,
			Error:       p.Error,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "JSON encode error: %v\n", err)
		os.Exit(1)
	}
}

func excerpt(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
