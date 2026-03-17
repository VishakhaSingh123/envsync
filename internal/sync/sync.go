package sync

import (
	"fmt"
	"strings"

	"github.com/envsync/internal/comparator"
	"github.com/envsync/internal/parser"
	"github.com/fatih/color"
)

// SyncPlan holds the final set of key-value pairs to write to the target
type SyncPlan struct {
	Changes map[string]string // key → new value
	Removes []string          // keys to remove from target (extra keys)
}

// BuildPlan constructs a SyncPlan from a drift report.
// If overwrite is false, conflicts prompt the user interactively.
func BuildPlan(
	src, tgt map[string]string,
	report *comparator.DriftReport,
	keysFilter string,
	overwrite bool,
) (*SyncPlan, error) {
	plan := &SyncPlan{Changes: make(map[string]string)}

	// Optional key filter
	filter := parseFilter(keysFilter)

	for _, entry := range report.GetEntries() {
		key := entry.GetKey()

		// Apply filter if specified
		if len(filter) > 0 {
			if _, wanted := filter[key]; !wanted {
				continue
			}
		}

		switch entry.GetStatus() {
		case comparator.GetStatusMissing():
			// Key in source, not in target → add it
			plan.Changes[key] = entry.GetSourceValue()

		case comparator.GetStatusMismatch():
			// Key differs → ask or overwrite
			if overwrite {
				plan.Changes[key] = entry.GetSourceValue()
			} else {
				val, err := resolveConflict(key, entry.GetSourceValue(), entry.GetTargetValue())
				if err != nil {
					return nil, err
				}
				if val != "" {
					plan.Changes[key] = val
				}
			}

		case comparator.GetStatusExtra():
			// Key in target but not source — leave as is (don't remove by default)
		}
	}

	return plan, nil
}

func resolveConflict(key, srcVal, tgtVal string) (string, error) {
	fmt.Println()
	color.Yellow("⚡ CONFLICT: %s\n", key)
	color.Cyan("  [S] Source value: %s\n", maskIfSensitive(key, srcVal))
	color.Blue("  [T] Target value: %s\n", maskIfSensitive(key, tgtVal))
	color.White("  [K] Keep target (skip)\n")

	fmt.Print("Choose [S/T/K]: ")
	var choice string
	fmt.Scanln(&choice)

	switch strings.ToUpper(strings.TrimSpace(choice)) {
	case "S":
		return srcVal, nil
	case "T":
		return tgtVal, nil
	case "K", "":
		return "", nil // skip
	default:
		return "", nil
	}
}

func maskIfSensitive(key, val string) string {
	lower := strings.ToLower(key)
	sensitiveWords := []string{"password", "secret", "token", "key", "private", "credential"}
	for _, w := range sensitiveWords {
		if strings.Contains(lower, w) {
			if len(val) <= 4 {
				return "****"
			}
			return val[:2] + strings.Repeat("*", len(val)-4) + val[len(val)-2:]
		}
	}
	return val
}

func parseFilter(keysFilter string) map[string]struct{} {
	result := make(map[string]struct{})
	if keysFilter == "" {
		return result
	}
	for _, k := range strings.Split(keysFilter, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			result[k] = struct{}{}
		}
	}
	return result
}

// Apply writes the plan to the target environment
func Apply(cfg *parser.Config, targetName string, plan *SyncPlan) (int, error) {
	// Load current target state
	current, err := parser.LoadEnvironment(cfg, targetName)
	if err != nil {
		return 0, fmt.Errorf("could not load target environment: %w", err)
	}

	// Merge changes
	for k, v := range plan.Changes {
		current[k] = v
	}

	// Write back
	if err := parser.WriteEnvironment(cfg, targetName, current); err != nil {
		return 0, fmt.Errorf("could not write to target environment: %w", err)
	}

	return len(plan.Changes), nil
}
