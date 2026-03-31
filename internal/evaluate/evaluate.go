// Package evaluate provides a comprehensive project completion evaluation system.
// It scans the codebase, cross-references specs, runs build/test checks, and
// produces a scored completion report with weighted categories.
package evaluate

import (
	"time"
)

// ProjectEvaluation is the top-level result of a project completion analysis.
type ProjectEvaluation struct {
	Timestamp    time.Time   `json:"timestamp"`
	OverallScore float64     `json:"overall_score"`
	Categories   []Category  `json:"categories"`
	Milestones   []Milestone `json:"milestones"`
	Summary      Summary     `json:"summary"`
}

// Category represents a scored evaluation area with weighted contribution to the overall score.
type Category struct {
	Name       string  `json:"name"`
	Weight     float64 `json:"weight"`
	Score      float64 `json:"score"`
	Percentage float64 `json:"percentage"`
	Items      []Item  `json:"items"`
}

// Item is a single pass/fail/partial check within a category.
type Item struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"` // "pass", "fail", "partial"
	Score    float64 `json:"score"`
	MaxScore float64 `json:"max_score"`
	Detail   string  `json:"detail"`
}

// Milestone tracks a user story's completion against its acceptance criteria.
type Milestone struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Priority  string   `json:"priority"`
	Status    string   `json:"status"` // "complete", "partial", "incomplete"
	Criteria  int      `json:"criteria"`
	Completed int      `json:"completed"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// Summary provides aggregate project statistics.
type Summary struct {
	TotalFiles      int    `json:"total_files"`
	TotalLOC        int    `json:"total_loc"`
	GoTests         int    `json:"go_tests"`
	FrontendTests   int    `json:"frontend_tests"`
	PassingTests    int    `json:"passing_tests"`
	FailingTests    int    `json:"failing_tests"`
	BuildStatus     string `json:"build_status"`
	TypeCheckStatus string `json:"typecheck_status"`
}

// Options configures the evaluation behavior.
type Options struct {
	SkipBuild    bool // Skip running go build / tsc / vite build
	SkipTests    bool // Skip running test suites
	SpecDir      string // Override spec directory (default: "spec")
	FrontendDir  string // Override frontend directory (default: "frontend")
}

// Evaluate runs a full project completion analysis on the given root directory.
func Evaluate(rootDir string, opts Options) (*ProjectEvaluation, error) {
	if opts.SpecDir == "" {
		opts.SpecDir = "spec"
	}
	if opts.FrontendDir == "" {
		opts.FrontendDir = "frontend"
	}

	eval := &ProjectEvaluation{
		Timestamp: time.Now().UTC(),
	}

	// Run all category checkers.
	eval.Categories = []Category{
		checkCodeImplementation(rootDir, opts),
		checkTestCoverage(rootDir, opts),
		checkSpecCompliance(rootDir, opts),
		checkBuildCI(rootDir, opts),
		checkDocumentation(rootDir, opts),
		checkSecurity(rootDir),
		checkInfrastructure(rootDir),
	}

	// Calculate overall weighted score.
	var totalWeighted float64
	for i := range eval.Categories {
		cat := &eval.Categories[i]
		// Calculate category percentage from items.
		var earned, max float64
		for _, item := range cat.Items {
			earned += item.Score
			max += item.MaxScore
		}
		if max > 0 {
			cat.Score = earned
			cat.Percentage = (earned / max) * 100
		}
		totalWeighted += cat.Percentage * cat.Weight
	}
	eval.OverallScore = totalWeighted

	// Parse and evaluate milestones from user stories.
	milestones, err := parseMilestones(rootDir, opts.SpecDir)
	if err == nil {
		eval.Milestones = milestones
	}

	// Build summary.
	eval.Summary = buildSummary(rootDir, opts)

	return eval, nil
}

func buildSummary(rootDir string, opts Options) Summary {
	goFiles := countFiles(rootDir, "*.go", []string{"node_modules", "vendor"})
	tsFiles := countFiles(rootDir, "*.ts", []string{"node_modules", "dist"})
	tsxFiles := countFiles(rootDir, "*.tsx", []string{"node_modules", "dist"})

	s := Summary{
		TotalFiles: goFiles + tsFiles + tsxFiles,
		TotalLOC:   countLOC(rootDir, []string{".go", ".ts", ".tsx"}, []string{"node_modules", "dist", "vendor"}),
	}

	goTestCount, _ := countTestFunctions(rootDir)
	s.GoTests = goTestCount

	// Count frontend tests from vitest output or test files.
	frontendTestCount := countFrontendTestFunctions(rootDir, opts.FrontendDir)
	s.FrontendTests = frontendTestCount

	if !opts.SkipBuild {
		if err := runGoBuild(rootDir); err != nil {
			s.BuildStatus = "fail"
		} else {
			s.BuildStatus = "pass"
		}
		if err := runTypeCheck(rootDir, opts.FrontendDir); err != nil {
			s.TypeCheckStatus = "fail"
		} else {
			s.TypeCheckStatus = "pass"
		}
	} else {
		s.BuildStatus = "skipped"
		s.TypeCheckStatus = "skipped"
	}

	if !opts.SkipTests {
		goPassed, goFailed, _ := runGoTests(rootDir)
		fePassed, feFailed, _ := runFrontendTests(rootDir, opts.FrontendDir)
		s.PassingTests = goPassed + fePassed
		s.FailingTests = goFailed + feFailed
	}

	return s
}

// CalculateOverallScore computes the weighted score from categories.
// Exported for testing.
func CalculateOverallScore(categories []Category) float64 {
	var total float64
	for _, cat := range categories {
		total += cat.Percentage * cat.Weight
	}
	return total
}
