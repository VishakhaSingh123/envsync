package cmd

import (
	"fmt"

	"github.com/envsync/internal/parser"
	"github.com/envsync/internal/snapshot"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage environment snapshots",
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [env]",
	Short: "Create a snapshot of an environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		envName := args[0]
		configPath, _ := cmd.Root().PersistentFlags().GetString("config")

		printBanner(fmt.Sprintf("Snapshot: %s", envName))

		cfg, err := parser.LoadConfig(configPath)
		if err != nil {
			exitErr("Failed to load config", err)
		}

		snap, err := snapshot.Create(cfg, envName)
		if err != nil {
			exitErr("Failed to create snapshot", err)
		}

		success(fmt.Sprintf("Snapshot created: %s (ID: %s)", envName, snap.ID))
		info(fmt.Sprintf("Saved to: %s", snap.Path))
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list [env]",
	Short: "List all snapshots for an environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		envName := args[0]
		configPath, _ := cmd.Root().PersistentFlags().GetString("config")

		cfg, err := parser.LoadConfig(configPath)
		if err != nil {
			exitErr("Failed to load config", err)
		}

		snaps, err := snapshot.List(cfg, envName)
		if err != nil {
			exitErr("Failed to list snapshots", err)
		}

		printBanner(fmt.Sprintf("Snapshots for: %s", envName))

		if len(snaps) == 0 {
			info("No snapshots found.")
			return
		}

		for i, s := range snaps {
			marker := "  "
			if i == 0 {
				marker = color.GreenString("▶ ")
			}
			fmt.Printf("%s[%s] %s — %d keys\n", marker, s.ID, s.CreatedAt.Format("2006-01-02 15:04:05"), s.KeyCount)
		}
	},
}

func init() {
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
}
