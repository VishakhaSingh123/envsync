package cmd

import (
	"fmt"
	"os"

	"github.com/envsync/internal/comparator"
	"github.com/envsync/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [env1] [env2]",
	Short: "Compare two environments and show drift report",
	Long: `Compares two environments and produces a detailed Drift Report.

Examples:
  envsync diff dev staging
  envsync diff staging production --strict
  envsync diff dev production --output json`,
	Args: cobra.ExactArgs(2),
	Run:  runDiff,
}

var diffOutputFormat string

func init() {
	diffCmd.Flags().StringVarP(&diffOutputFormat, "output", "o", "table", "Output format: table|json|yaml")
}

func runDiff(cmd *cobra.Command, args []string) {
	env1Name, env2Name := args[0], args[1]
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	printBanner(fmt.Sprintf("Drift Report: %s → %s", env1Name, env2Name))

	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		exitErr("Failed to load envsync config", err)
	}

	env1, err := parser.LoadEnvironment(cfg, env1Name)
	if err != nil {
		exitErr(fmt.Sprintf("Failed to load environment '%s'", env1Name), err)
	}

	env2, err := parser.LoadEnvironment(cfg, env2Name)
	if err != nil {
		exitErr(fmt.Sprintf("Failed to load environment '%s'", env2Name), err)
	}

	report := comparator.Compare(env1, env2)

	switch diffOutputFormat {
	case "json":
		comparator.PrintJSON(report, os.Stdout)
	case "yaml":
		comparator.PrintYAML(report, os.Stdout)
	default:
		comparator.PrintTable(report, env1Name, env2Name)
	}

	if report.HasDrift() {
		color.Yellow("\n⚠  Drift detected: %d missing, %d mismatched, %d extra\n",
			report.MissingCount(), report.MismatchCount(), report.ExtraCount())
		os.Exit(2) // Exit code 2 = drift detected (useful for CI)
	} else {
		color.Green("\n✔  Environments are in sync!\n")
	}
}
