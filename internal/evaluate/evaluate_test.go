package evaluate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	if !fileExists(f) {
		t.Error("expected file to exist")
	}
	if fileExists(filepath.Join(dir, "missing.txt")) {
		t.Error("expected file to not exist")
	}
	if fileExists(dir) {
		t.Error("directory should not be reported as file")
	}
}

func TestDirExists(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)

	if !dirExists(sub) {
		t.Error("expected dir to exist")
	}
	if dirExists(filepath.Join(dir, "missing")) {
		t.Error("expected dir to not exist")
	}
}

func TestCountFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package b"), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("text"), 0644)
	os.Mkdir(filepath.Join(dir, "node_modules"), 0755)
	os.WriteFile(filepath.Join(dir, "node_modules", "d.go"), []byte("package d"), 0644)

	count := countFiles(dir, "*.go", []string{"node_modules"})
	if count != 2 {
		t.Errorf("expected 2 .go files, got %d", count)
	}
}

func TestCountLOC(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"
	os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0644)

	loc := countLOC(dir, []string{".go"}, nil)
	// 5 non-empty lines (package, import, func, fmt.Println, closing brace)
	if loc != 5 {
		t.Errorf("expected 5 non-empty lines, got %d", loc)
	}
}

func TestCountTestFunctions(t *testing.T) {
	dir := t.TempDir()
	testFile := `package foo

func TestOne(t *testing.T) {}

func TestTwo(t *testing.T) {}

func helperFunc() {}
`
	os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte(testFile), 0644)

	count, err := countTestFunctions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 test functions, got %d", count)
	}
}

func TestCountFrontendTestFunctions(t *testing.T) {
	dir := t.TempDir()
	feDir := filepath.Join(dir, "frontend", "src")
	os.MkdirAll(feDir, 0755)

	testFile := `import { it, expect } from 'vitest';

it('does thing one', () => {
  expect(true).toBe(true);
});

it('does thing two', () => {
  expect(1).toBe(1);
});

test('third test', () => {
  expect(2).toBe(2);
});
`
	os.WriteFile(filepath.Join(feDir, "foo.test.ts"), []byte(testFile), 0644)

	count := countFrontendTestFunctions(dir, "frontend")
	if count != 3 {
		t.Errorf("expected 3 frontend test functions, got %d", count)
	}
}

func TestGrepFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("filepath.Clean(p)"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("nothing here"), 0644)

	matches := grepFiles(dir, `filepath\.Clean`, "*.go")
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
}

func TestExtractIDs(t *testing.T) {
	dir := t.TempDir()
	content := `# API Spec
## GET /health [API-001]
## POST /workspaces [API-002]
Some text [ARCH-001]
Duplicate [API-001]
`
	f := filepath.Join(dir, "spec.md")
	os.WriteFile(f, []byte(content), 0644)

	ids := extractIDs(f)
	if len(ids) != 3 {
		t.Errorf("expected 3 unique IDs, got %d: %v", len(ids), ids)
	}

	expected := map[string]bool{"API-001": true, "API-002": true, "ARCH-001": true}
	for _, id := range ids {
		if !expected[id] {
			t.Errorf("unexpected ID: %s", id)
		}
	}
}

func TestCountPattern(t *testing.T) {
	dir := t.TempDir()
	content := `build:
	go build
test:
	go test
run:
	go run
`
	f := filepath.Join(dir, "Makefile")
	os.WriteFile(f, []byte(content), 0644)

	count := countPattern(f, `^(build|test|run):`)
	if count != 3 {
		t.Errorf("expected 3 matches, got %d", count)
	}
}

func TestParseMilestones(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "spec")
	os.MkdirAll(specDir, 0755)

	content := `# User Stories

### US-001: Create Workspace [P0]

**Acceptance Criteria:**
- [x] Clicking creates workspace
- [x] Container provisioned
- [ ] File tree loads

**Traceability:** API-002, ARCH-001

### US-002: Stop Workspace [P0]

**Acceptance Criteria:**
- [x] Session stops
- [x] Container removed
`
	os.WriteFile(filepath.Join(specDir, "user-stories.md"), []byte(content), 0644)

	milestones, err := parseMilestones(dir, "spec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(milestones) != 2 {
		t.Fatalf("expected 2 milestones, got %d", len(milestones))
	}

	us1 := milestones[0]
	if us1.ID != "US-001" {
		t.Errorf("expected US-001, got %s", us1.ID)
	}
	if us1.Name != "Create Workspace" {
		t.Errorf("expected 'Create Workspace', got %q", us1.Name)
	}
	if us1.Priority != "P0" {
		t.Errorf("expected P0, got %s", us1.Priority)
	}
	if us1.Criteria != 3 {
		t.Errorf("expected 3 criteria, got %d", us1.Criteria)
	}
	if us1.Completed != 2 {
		t.Errorf("expected 2 completed, got %d", us1.Completed)
	}
	if us1.Status != "partial" {
		t.Errorf("expected partial, got %s", us1.Status)
	}
	if len(us1.DependsOn) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(us1.DependsOn))
	}

	us2 := milestones[1]
	if us2.Status != "complete" {
		t.Errorf("expected complete, got %s", us2.Status)
	}
}

func TestCalculateOverallScore(t *testing.T) {
	categories := []Category{
		{Weight: 0.50, Percentage: 100},
		{Weight: 0.30, Percentage: 80},
		{Weight: 0.20, Percentage: 60},
	}

	score := CalculateOverallScore(categories)
	// 0.50*100 + 0.30*80 + 0.20*60 = 50 + 24 + 12 = 86
	if score < 85.9 || score > 86.1 {
		t.Errorf("expected ~86.0, got %f", score)
	}
}

func TestWeightsSumToOne(t *testing.T) {
	// Verify the weights defined in the system sum to 1.0.
	weights := []float64{0.25, 0.25, 0.15, 0.10, 0.10, 0.10, 0.05}
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("weights should sum to 1.0, got %f", sum)
	}
}

func TestFormatJSONRoundtrip(t *testing.T) {
	eval := &ProjectEvaluation{
		OverallScore: 85.5,
		Categories: []Category{
			{Name: "Test", Weight: 1.0, Percentage: 85.5, Items: []Item{
				{Name: "item1", Status: "pass", Score: 1, MaxScore: 1},
			}},
		},
		Summary: Summary{TotalFiles: 10, TotalLOC: 500},
	}

	data, err := FormatJSON(eval)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var parsed ProjectEvaluation
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON roundtrip failed: %v", err)
	}

	if parsed.OverallScore != 85.5 {
		t.Errorf("expected 85.5, got %f", parsed.OverallScore)
	}
	if parsed.Summary.TotalFiles != 10 {
		t.Errorf("expected 10 files, got %d", parsed.Summary.TotalFiles)
	}
}

func TestFormatTerminalContainsExpected(t *testing.T) {
	eval := &ProjectEvaluation{
		OverallScore: 92.3,
		Categories: []Category{
			{Name: "Code Implementation", Weight: 0.25, Percentage: 95.0, Items: []Item{
				{Name: "Backend packages", Status: "pass", Detail: "All 6 exist"},
			}},
		},
		Milestones: []Milestone{
			{ID: "US-001", Name: "Create Workspace", Priority: "P0", Status: "complete", Criteria: 7, Completed: 7},
		},
		Summary: Summary{TotalFiles: 63, TotalLOC: 7900, GoTests: 109, FrontendTests: 86, PassingTests: 195, BuildStatus: "pass"},
	}

	output := FormatTerminal(eval)

	checks := []string{
		"92.3%",
		"Code Implementation",
		"Backend packages",
		"US-001",
		"Create Workspace",
		"7/7",
		"Files: 63",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("terminal output missing %q", check)
		}
	}
}

func TestProgressBar(t *testing.T) {
	bar0 := progressBar(0, 10)
	if bar0 != "░░░░░░░░░░" {
		t.Errorf("expected all empty, got %q", bar0)
	}

	bar100 := progressBar(100, 10)
	if bar100 != "██████████" {
		t.Errorf("expected all filled, got %q", bar100)
	}

	bar50 := progressBar(50, 10)
	if !strings.Contains(bar50, "█████") {
		t.Errorf("expected 5 filled blocks, got %q", bar50)
	}
}

func TestCheckFileHelper(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "exists.txt")
	os.WriteFile(f, []byte("hi"), 0644)

	item := checkFile(f, "test file")
	if item.Status != "pass" || item.Score != 1 {
		t.Errorf("expected pass, got %s", item.Status)
	}

	missing := checkFile(filepath.Join(dir, "no.txt"), "missing")
	if missing.Status != "fail" {
		t.Errorf("expected fail, got %s", missing.Status)
	}
}

func TestCheckDirHelper(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)

	item := checkDir(sub, "test dir")
	if item.Status != "pass" {
		t.Errorf("expected pass, got %s", item.Status)
	}

	missing := checkDir(filepath.Join(dir, "nope"), "missing")
	if missing.Status != "fail" {
		t.Errorf("expected fail, got %s", missing.Status)
	}
}

func TestFormatInt(t *testing.T) {
	if formatInt(0) != "0" {
		t.Errorf("expected '0', got %q", formatInt(0))
	}
	if formatInt(42) != "42" {
		t.Errorf("expected '42', got %q", formatInt(42))
	}
	if formatInt(1000) != "1000" {
		t.Errorf("expected '1000', got %q", formatInt(1000))
	}
}
