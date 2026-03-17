package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envsync",
	Short: "EnvSync — Environment Synchronization Tool",
	Long: color.New(color.FgCyan, color.Bold).Sprint(`
╔═══════════════════════════════════════════════════════╗
║              EnvSync v1.0 — Config Parity Tool        ║
║   Detect drift • Sync secrets • Validate runtimes     ║
╚═══════════════════════════════════════════════════════╝
`) + `
EnvSync audits and synchronizes environment variables,
infrastructure state, and config files across Dev → Staging → Production.

Usage examples:
  envsync diff dev staging           Diff two environments
  envsync sync dev staging           Sync dev → staging
  envsync audit --env staging        Audit a single environment
  envsync snapshot create staging    Snapshot staging state
  envsync rollback staging           Rollback staging to last snapshot
  envsync validate --env dev         Validate runtime versions
`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "envsync.yaml", "Path to envsync config file")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("strict", false, "Strict mode: production requires PR-style approval")
}

func exitErr(msg string, err error) {
	color.Red("✗ %s: %v\n", msg, err)
	os.Exit(1)
}

func success(msg string) {
	color.Green("✔ %s\n", msg)
}

func info(msg string) {
	color.Cyan("ℹ %s\n", msg)
}

func warn(msg string) {
	color.Yellow("⚠ %s\n", msg)
}

func printBanner(title string) {
	c := color.New(color.FgCyan, color.Bold)
	fmt.Println()
	c.Printf("══ %s ══\n\n", title)
}

func confirmPrompt(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	var answer string
	fmt.Scanln(&answer)
	return answer == "y" || answer == "Y"
}
