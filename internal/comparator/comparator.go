package comparator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// ── Types ─────────────────────────────────────────────────────────────────────

type DiffStatus string

const (
	StatusMissing  DiffStatus = "MISSING"   // in source but not in target
	StatusMismatch DiffStatus = "MISMATCH"  // in both but different value
	StatusExtra    DiffStatus = "EXTRA"     // in target but not in source
	StatusMatch    DiffStatus = "MATCH"     // identical
)

type DiffEntry struct {
	Key         string     `json:"key"`
	Status      DiffStatus `json:"status"`
	SourceValue string     `json:"source_value,omitempty"`
	TargetValue string     `json:"target_value,omitempty"`
}

type DriftReport struct {
	Entries []DiffEntry `json:"entries"`
}

// ── Core comparison ───────────────────────────────────────────────────────────

var sensitivePatterns = []string{
	"password", "secret", "token", "key", "private", "credential",
	"auth", "jwt", "api_key", "apikey",
}

func isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, p := range sensitivePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func maskValue(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return val[:2] + strings.Repeat("*", len(val)-4) + val[len(val)-2:]
}

func displayValue(key, val string) string {
	if val == "" {
		return color.YellowString("(empty)")
	}
	if isSensitive(key) {
		return color.MagentaString(maskValue(val))
	}
	return val
}

// Compare produces a DriftReport between two environments.
// source is the "desired state", target is what we're comparing against.
func Compare(source, target map[string]string) *DriftReport {
	report := &DriftReport{}

	// Keys in source
	for key, srcVal := range source {
		tgtVal, exists := target[key]
		if !exists {
			report.Entries = append(report.Entries, DiffEntry{
				Key:         key,
				Status:      StatusMissing,
				SourceValue: srcVal,
			})
		} else if srcVal != tgtVal {
			report.Entries = append(report.Entries, DiffEntry{
				Key:         key,
				Status:      StatusMismatch,
				SourceValue: srcVal,
				TargetValue: tgtVal,
			})
		} else {
			report.Entries = append(report.Entries, DiffEntry{
				Key:    key,
				Status: StatusMatch,
			})
		}
	}

	// Extra keys in target
	for key, tgtVal := range target {
		if _, exists := source[key]; !exists {
			report.Entries = append(report.Entries, DiffEntry{
				Key:         key,
				Status:      StatusExtra,
				TargetValue: tgtVal,
			})
		}
	}

	// Sort by status priority then key
	sort.Slice(report.Entries, func(i, j int) bool {
		pi := statusPriority(report.Entries[i].Status)
		pj := statusPriority(report.Entries[j].Status)
		if pi != pj {
			return pi < pj
		}
		return report.Entries[i].Key < report.Entries[j].Key
	})

	return report
}

func statusPriority(s DiffStatus) int {
	switch s {
	case StatusMissing:
		return 0
	case StatusMismatch:
		return 1
	case StatusExtra:
		return 2
	default:
		return 3
	}
}

// ── Counts ────────────────────────────────────────────────────────────────────

func (r *DriftReport) HasDrift() bool {
	return r.MissingCount()+r.MismatchCount()+r.ExtraCount() > 0
}

func (r *DriftReport) MissingCount() int  { return r.countByStatus(StatusMissing) }
func (r *DriftReport) MismatchCount() int { return r.countByStatus(StatusMismatch) }
func (r *DriftReport) ExtraCount() int    { return r.countByStatus(StatusExtra) }

func (r *DriftReport) countByStatus(s DiffStatus) int {
	n := 0
	for _, e := range r.Entries {
		if e.Status == s {
			n++
		}
	}
	return n
}

// ── Printers ─────────────────────────────────────────────────────────────────

func PrintTable(report *DriftReport, srcName, tgtName string) {
	colW := 32
	valW := 28

	header := fmt.Sprintf("  %-*s  %-10s  %-*s  %-*s",
		colW, "KEY", "STATUS", valW, srcName, valW, tgtName)
	fmt.Println(color.New(color.FgWhite, color.Bold).Sprint(header))
	fmt.Println(strings.Repeat("─", len(header)+2))

	for _, e := range report.Entries {
		if e.Status == StatusMatch {
			continue // Only show drifted entries
		}

		key := truncate(e.Key, colW)
		src := truncate(displayValue(e.Key, e.SourceValue), valW)
		tgt := truncate(displayValue(e.Key, e.TargetValue), valW)

		var statusStr string
		switch e.Status {
		case StatusMissing:
			statusStr = color.RedString("%-10s", "MISSING")
		case StatusMismatch:
			statusStr = color.YellowString("%-10s", "MISMATCH")
		case StatusExtra:
			statusStr = color.BlueString("%-10s", "EXTRA")
		}

		fmt.Printf("  %-*s  %s  %-*s  %-*s\n",
			colW, key, statusStr, valW, src, valW, tgt)
	}

	// Summary of matches
	matches := 0
	for _, e := range report.Entries {
		if e.Status == StatusMatch {
			matches++
		}
	}
	if matches > 0 {
		fmt.Printf(color.GreenString("\n  + %d keys match (hidden)\n"), matches)
	}
}

func PrintJSON(report *DriftReport, w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(report)
}

func PrintYAML(report *DriftReport, w io.Writer) {
	enc := yaml.NewEncoder(w)
	enc.Encode(report)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Exported for sync package
func (r *DriftReport) GetEntries() []DiffEntry { return r.Entries }
func (e DiffEntry) GetKey() string             { return e.Key }
func (e DiffEntry) GetStatus() DiffStatus      { return e.Status }
func (e DiffEntry) GetSourceValue() string     { return e.SourceValue }
func (e DiffEntry) GetTargetValue() string     { return e.TargetValue }
func GetStatusMissing() DiffStatus             { return StatusMissing }
func GetStatusMismatch() DiffStatus            { return StatusMismatch }
func GetStatusExtra() DiffStatus               { return StatusExtra }

// SaveJSON writes the report to a file (used by CI)
func (r *DriftReport) SaveJSON(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	PrintJSON(r, f)
	return nil
}
