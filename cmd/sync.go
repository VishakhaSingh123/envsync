package cmd

import (
	"fmt"

	"github.com/envsync/internal/comparator"
	"github.com/envsync/internal/parser"
	"github.com/envsync/internal/snapshot"
	"github.com/envsync/internal/sync"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [source] [target]",
	Short: "Synchronize source environment into target",
	Long: `Push configuration from source environment to target.
Automatically snapshots target before applying changes.

Examples:
  envsync sync dev staging
  envsync sync staging production --strict
  envsync sync dev staging --dry-run
  envsync sync dev staging --keys DB_HOST,REDIS_URL`,
	Args: cobra.ExactArgs(2),
	Run:  runSync,
}

var (
	syncDryRun    bool
	syncKeys      string
	syncOverwrite bool
)

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without applying them")
	syncCmd.Flags().StringVar(&syncKeys, "keys", "", "Comma-separated list of specific keys to sync")
	syncCmd.Flags().BoolVar(&syncOverwrite, "overwrite", false, "Overwrite conflicting values without prompting")
}

func runSync(cmd *cobra.Command, args []string) {
	srcName, tgtName := args[0], args[1]
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")
	strictMode, _ := cmd.Root().PersistentFlags().GetBool("strict")

	printBanner(fmt.Sprintf("Sync: %s → %s", srcName, tgtName))

	// Strict mode: production requires explicit confirmation
	if strictMode && (tgtName == "production" || tgtName == "prod") {
		warn("STRICT MODE: Syncing to production requires approval.")
		if !confirmPrompt("Are you sure you want to sync to PRODUCTION?") {
			info("Sync cancelled.")
			return
		}
	}

	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		exitErr("Failed to load config", err)
	}

	src, err := parser.LoadEnvironment(cfg, srcName)
	if err != nil {
		exitErr(fmt.Sprintf("Failed to load source '%s'", srcName), err)
	}

	tgt, err := parser.LoadEnvironment(cfg, tgtName)
	if err != nil {
		exitErr(fmt.Sprintf("Failed to load target '%s'", tgtName), err)
	}

	// Diff first
	report := comparator.Compare(src, tgt)

	if !report.HasDrift() {
		success("Environments are already in sync. Nothing to do.")
		return
	}

	comparator.PrintTable(report, srcName, tgtName)

	if syncDryRun {
		info("DRY RUN: No changes applied.")
		return
	}

	// Auto-snapshot before sync
	info(fmt.Sprintf("Creating snapshot of '%s' before sync...", tgtName))
	snap, err := snapshot.Create(cfg, tgtName)
	if err != nil {
		warn(fmt.Sprintf("Could not create snapshot: %v (continuing anyway)", err))
	} else {
		success(fmt.Sprintf("Snapshot saved: %s", snap.ID))
	}

	// Resolve conflicts
	plan, err := sync.BuildPlan(src, tgt, report, syncKeys, syncOverwrite)
	if err != nil {
		exitErr("Failed to build sync plan", err)
	}

	// Apply
	applied, err := sync.Apply(cfg, tgtName, plan)
	if err != nil {
		exitErr("Sync failed", err)
		color.Red("Run: envsync rollback %s  to restore previous state\n", tgtName)
		return
	}

	fmt.Println()
	success(fmt.Sprintf("Sync complete: %d keys applied to '%s'", applied, tgtName))
}
