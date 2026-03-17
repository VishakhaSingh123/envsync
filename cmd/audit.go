package cmd

import (
	"fmt"
	"os"

	"github.com/envsync/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit a single environment for missing/undefined keys",
	Long: `Audit compares an environment against the .env.example source of truth.
Useful as a pre-deployment check in CI/CD pipelines.

Examples:
  envsync audit --env staging
  envsync audit --env production --fail-on-missing
  envsync audit --env dev --threshold 10`,
	Run: runAudit,
}

var (
	auditEnv          string
	auditFailOnMissing bool
	auditThreshold    int
)

func init() {
	auditCmd.Flags().StringVar(&auditEnv, "env", "dev", "Environment to audit")
	auditCmd.Flags().BoolVar(&auditFailOnMissing, "fail-on-missing", false, "Exit with code 1 if any keys are missing")
	auditCmd.Flags().IntVar(&auditThreshold, "threshold", 0, "Max allowed drift count (0 = unlimited)")
}

func runAudit(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	printBanner(fmt.Sprintf("Audit: %s", auditEnv))

	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		exitErr("Failed to load config", err)
	}

	env, err := parser.LoadEnvironment(cfg, auditEnv)
	if err != nil {
		exitErr(fmt.Sprintf("Failed to load environment '%s'", auditEnv), err)
	}

	truth, err := parser.LoadSourceOfTruth(cfg)
	if err != nil {
		exitErr("Failed to load source of truth (.env.example)", err)
	}

	missing := []string{}
	empty := []string{}
	extra := []string{}
	ok := 0

	for key := range truth {
		val, exists := env[key]
		if !exists {
			missing = append(missing, key)
		} else if val == "" || val == "CHANGE_ME" || val == "TODO" {
			empty = append(empty, key)
		} else {
			ok++
		}
	}

	for key := range env {
		if _, exists := truth[key]; !exists {
			extra = append(extra, key)
		}
	}

	// Print results
	if len(missing) > 0 {
		color.Red("\n✗ MISSING keys (not in %s):\n", auditEnv)
		for _, k := range missing {
			color.Red("   - %s\n", k)
		}
	}

	if len(empty) > 0 {
		color.Yellow("\n⚠ EMPTY / PLACEHOLDER keys:\n")
		for _, k := range empty {
			color.Yellow("   ~ %s\n", k)
		}
	}

	if len(extra) > 0 {
		color.Blue("\nℹ EXTRA keys (not in .env.example):\n")
		for _, k := range extra {
			color.Blue("   + %s\n", k)
		}
	}

	fmt.Println()
	color.Green("✔ OK: %d keys\n", ok)
	color.Red("✗ Missing: %d keys\n", len(missing))
	color.Yellow("~ Empty: %d keys\n", len(empty))
	color.Blue("+ Extra: %d keys\n", len(extra))

	totalDrift := len(missing) + len(empty)

	if auditThreshold > 0 && totalDrift > auditThreshold {
		color.Red("\n✗ AUDIT FAILED: Drift count %d exceeds threshold %d\n", totalDrift, auditThreshold)
		os.Exit(1)
	}

	if auditFailOnMissing && len(missing) > 0 {
		color.Red("\n✗ AUDIT FAILED: Missing keys detected\n")
		os.Exit(1)
	}

	if totalDrift == 0 {
		success(fmt.Sprintf("Environment '%s' is clean ✓", auditEnv))
	} else {
		warn(fmt.Sprintf("Environment '%s' has %d issues", auditEnv, totalDrift))
	}
}
