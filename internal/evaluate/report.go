package evaluate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON returns the evaluation report as indented JSON.
func FormatJSON(eval *ProjectEvaluation) ([]byte, error) {
	return json.MarshalIndent(eval, "", "  ")
}

// FormatTerminal returns a human-readable terminal report with progress bars and status indicators.
func FormatTerminal(eval *ProjectEvaluation) string {
	var b strings.Builder

	// Header.
	b.WriteString("\n")
	b.WriteString("╔══════════════════════════════════════════════════════╗\n")
	b.WriteString(fmt.Sprintf("║        CloudCode Project Evaluation                 ║\n"))
	b.WriteString(fmt.Sprintf("║        Overall Completion: %5.1f%%                    ║\n", eval.OverallScore))
	b.WriteString("╚══════════════════════════════════════════════════════╝\n")
	b.WriteString("\n")

	// Categories.
	for _, cat := range eval.Categories {
		bar := progressBar(cat.Percentage, 20)
		weightPct := cat.Weight * 100
		b.WriteString(fmt.Sprintf("%-24s [%s] %5.1f%%  (%.0f%% weight)\n", cat.Name, bar, cat.Percentage, weightPct))

		for _, item := range cat.Items {
			icon := statusIcon(item.Status)
			b.WriteString(fmt.Sprintf("  %s %s", icon, item.Name))
			if item.Detail != "" {
				b.WriteString(fmt.Sprintf(" — %s", item.Detail))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Milestones.
	if len(eval.Milestones) > 0 {
		b.WriteString("Milestones:\n")
		for _, m := range eval.Milestones {
			icon := milestoneIcon(m.Status)
			criteriaStr := fmt.Sprintf("%d/%d criteria", m.Completed, m.Criteria)
			if m.Status == "partial" {
				criteriaStr += " (partial)"
			}
			b.WriteString(fmt.Sprintf("  %s %-8s %-30s [%s]  %s\n", icon, m.ID, m.Name, m.Priority, criteriaStr))
		}
		b.WriteString("\n")
	}

	// Summary.
	b.WriteString("Summary:\n")
	b.WriteString(fmt.Sprintf("  Files: %d | LOC: %d | Tests: %d (%d passing, %d failing)\n",
		eval.Summary.TotalFiles, eval.Summary.TotalLOC,
		eval.Summary.GoTests+eval.Summary.FrontendTests,
		eval.Summary.PassingTests, eval.Summary.FailingTests))
	b.WriteString(fmt.Sprintf("  Build: %s | TypeCheck: %s\n",
		statusIcon(eval.Summary.BuildStatus), statusIcon(eval.Summary.TypeCheckStatus)))
	b.WriteString("\n")

	return b.String()
}

func progressBar(pct float64, width int) string {
	if pct > 100 {
		pct = 100
	}
	if pct < 0 {
		pct = 0
	}
	filled := int(pct / 100 * float64(width))
	empty := width - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

func statusIcon(status string) string {
	switch status {
	case "pass":
		return "✓"
	case "fail":
		return "✗"
	case "partial":
		return "~"
	case "skipped":
		return "-"
	default:
		return "?"
	}
}

func milestoneIcon(status string) string {
	switch status {
	case "complete":
		return "✓"
	case "partial":
		return "~"
	case "incomplete":
		return "✗"
	default:
		return "?"
	}
}
