package evaluate

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// parseMilestones reads spec/user-stories.md and extracts milestone data.
func parseMilestones(root, specDir string) ([]Milestone, error) {
	usPath := filepath.Join(root, specDir, "user-stories.md")
	f, err := os.Open(usPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var milestones []Milestone
	var current *Milestone

	// Patterns.
	headerRe := regexp.MustCompile(`^###\s+(\S+):\s+(.+?)\s+\[(P\d)\]`)
	criteriaChecked := regexp.MustCompile(`^-\s+\[x\]`)
	criteriaUnchecked := regexp.MustCompile(`^-\s+\[\s*\]`)
	traceRe := regexp.MustCompile(`\*\*Traceability:\*\*\s*(.+)`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Match user story header: ### US-001: Create Workspace [P0]
		if m := headerRe.FindStringSubmatch(line); len(m) == 4 {
			if current != nil {
				milestones = append(milestones, *current)
			}
			current = &Milestone{
				ID:       m[1],
				Name:     m[2],
				Priority: m[3],
			}
			continue
		}

		if current == nil {
			continue
		}

		// Count acceptance criteria.
		if criteriaChecked.MatchString(line) {
			current.Criteria++
			current.Completed++
		} else if criteriaUnchecked.MatchString(line) {
			current.Criteria++
		}

		// Parse traceability references as dependencies.
		if m := traceRe.FindStringSubmatch(line); len(m) == 2 {
			refs := strings.Split(m[1], ",")
			for _, ref := range refs {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					current.DependsOn = append(current.DependsOn, ref)
				}
			}
		}
	}

	// Don't forget the last milestone.
	if current != nil {
		milestones = append(milestones, *current)
	}

	// Set status for each milestone.
	for i := range milestones {
		m := &milestones[i]
		if m.Criteria == 0 {
			m.Status = "incomplete"
		} else if m.Completed == m.Criteria {
			m.Status = "complete"
		} else if m.Completed > 0 {
			m.Status = "partial"
		} else {
			m.Status = "incomplete"
		}
	}

	return milestones, nil
}
