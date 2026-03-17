package cmd

import (
	"fmt"

	"github.com/envsync/internal/parser"
	"github.com/envsync/internal/snapshot"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [env]",
	Short: "Rollback environment to the last snapshot",
	Long: `Restore an environment to its most recent snapshot.
Use --id to rollback to a specific snapshot.

Examples:
  envsync rollback staging
  envsync rollback production --id snap_20240312_143022`,
	Args: cobra.ExactArgs(1),
	Run:  runRollback,
}

var rollbackID string

func init() {
	rollbackCmd.Flags().StringVar(&rollbackID, "id", "", "Snapshot ID to rollback to (default: latest)")
}

func runRollback(cmd *cobra.Command, args []string) {
	envName := args[0]
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	printBanner(fmt.Sprintf("Rollback: %s", envName))

	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		exitErr("Failed to load config", err)
	}

	warn(fmt.Sprintf("This will overwrite the current state of '%s'.", envName))
	if !confirmPrompt("Proceed with rollback?") {
		info("Rollback cancelled.")
		return
	}

	snap, err := snapshot.Restore(cfg, envName, rollbackID)
	if err != nil {
		exitErr("Rollback failed", err)
	}

	success(fmt.Sprintf("Rollback complete: '%s' restored to snapshot %s", envName, snap.ID))
	info(fmt.Sprintf("Snapshot was taken at: %s", snap.CreatedAt.Format("2006-01-02 15:04:05")))
}
