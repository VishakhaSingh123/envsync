package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/envsync/internal/parser"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ── validate ──────────────────────────────────────────────────────────────────

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate runtime versions against required spec",
	Long: `Check that installed runtime versions match the versions defined
in envsync.yaml under the 'runtimes' section.

Examples:
  envsync validate --env dev
  envsync validate --env staging`,
	Run: runValidate,
}

var validateEnv string

func init() {
	validateCmd.Flags().StringVar(&validateEnv, "env", "dev", "Environment to validate")
}

func runValidate(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Root().PersistentFlags().GetString("config")

	printBanner(fmt.Sprintf("Runtime Validation: %s", validateEnv))

	cfg, err := parser.LoadConfig(configPath)
	if err != nil {
		exitErr("Failed to load config", err)
	}

	if cfg.Runtimes == nil || len(cfg.Runtimes) == 0 {
		info("No runtimes defined in envsync.yaml. Skipping.")
		return
	}

	allPass := true
	for runtime, required := range cfg.Runtimes {
		installed, err := getInstalledVersion(runtime)
		if err != nil {
			color.Red("  ✗ %-12s required: %-12s installed: not found\n", runtime, required)
			allPass = false
			continue
		}

		if versionMatches(installed, required) {
			color.Green("  ✔ %-12s required: %-12s installed: %s\n", runtime, required, installed)
		} else {
			color.Red("  ✗ %-12s required: %-12s installed: %s\n", runtime, required, installed)
			allPass = false
		}
	}

	fmt.Println()
	if allPass {
		success("All runtime versions match!")
	} else {
		color.Red("✗ Runtime version mismatch detected.\n")
	}
}

func getInstalledVersion(runtime string) (string, error) {
	var out []byte
	var err error
	switch runtime {
	case "node":
		out, err = exec.Command("node", "--version").Output()
	case "python", "python3":
		out, err = exec.Command("python3", "--version").Output()
	case "go":
		out, err = exec.Command("go", "version").Output()
	case "ruby":
		out, err = exec.Command("ruby", "--version").Output()
	case "java":
		out, err = exec.Command("java", "-version").CombinedOutput()
	default:
		out, err = exec.Command(runtime, "--version").Output()
	}
	if err != nil {
		return "", err
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	for _, p := range parts {
		if len(p) > 0 && (p[0] == 'v' || (p[0] >= '0' && p[0] <= '9')) {
			return strings.TrimPrefix(p, "v"), nil
		}
	}
	return strings.TrimSpace(string(out)), nil
}

func versionMatches(installed, required string) bool {
	installed = strings.TrimPrefix(installed, "v")
	required = strings.TrimPrefix(required, "v")
	// Prefix match: "20" matches "20.11.1"
	return strings.HasPrefix(installed, required)
}

// ── init ──────────────────────────────────────────────────────────────────────

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a new envsync.yaml config file",
	Run: func(cmd *cobra.Command, args []string) {
		printBanner("Init EnvSync Project")
		err := parser.ScaffoldConfig("envsync.yaml")
		if err != nil {
			exitErr("Failed to scaffold config", err)
		}
		success("Created envsync.yaml")
		info("Edit this file to point to your environment sources, then run: envsync audit --env dev")
	},
}
