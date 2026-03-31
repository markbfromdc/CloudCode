package evaluate

import (
	"path/filepath"
	"strconv"
	"strings"
)

// checkCodeImplementation verifies that all planned source files and packages exist (weight: 0.25).
func checkCodeImplementation(root string, opts Options) Category {
	cat := Category{Name: "Code Implementation", Weight: 0.25}

	// Backend packages.
	backendPkgs := []string{"api", "config", "container", "middleware", "websocket", "logging"}
	for _, pkg := range backendPkgs {
		path := filepath.Join(root, "internal", pkg)
		cat.Items = append(cat.Items, checkDir(path, "Backend package: internal/"+pkg))
	}

	// Server entry point.
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "cmd", "server", "main.go"), "Server entry: cmd/server/main.go"))

	// Frontend directories.
	feDirs := []string{"components", "services", "hooks", "context", "types"}
	for _, dir := range feDirs {
		path := filepath.Join(root, opts.FrontendDir, "src", dir)
		cat.Items = append(cat.Items, checkDir(path, "Frontend dir: src/"+dir))
	}

	// API endpoint registrations (count HandleFunc/Handle in main.go).
	mainPath := filepath.Join(root, "cmd", "server", "main.go")
	handleCount := countPattern(mainPath, `\.(HandleFunc|Handle)\(`)
	item := Item{Name: "API endpoint registrations", MaxScore: 1}
	if handleCount >= 16 {
		item.Status = "pass"
		item.Score = 1
		item.Detail = formatInt(handleCount) + " endpoints registered (target: 16)"
	} else if handleCount >= 10 {
		item.Status = "partial"
		item.Score = float64(handleCount) / 16.0
		item.Detail = formatInt(handleCount) + "/16 endpoints registered"
	} else {
		item.Status = "fail"
		item.Detail = formatInt(handleCount) + "/16 endpoints registered"
	}
	cat.Items = append(cat.Items, item)

	// Frontend component count.
	tsxCount := countFiles(filepath.Join(root, opts.FrontendDir, "src", "components"), "*.tsx", []string{"node_modules"})
	compItem := Item{Name: "Frontend components", MaxScore: 1}
	if tsxCount >= 18 {
		compItem.Status = "pass"
		compItem.Score = 1
		compItem.Detail = formatInt(tsxCount) + " components (target: 18)"
	} else {
		compItem.Status = "partial"
		compItem.Score = float64(tsxCount) / 18.0
		compItem.Detail = formatInt(tsxCount) + "/18 components"
	}
	cat.Items = append(cat.Items, compItem)

	return cat
}

// checkTestCoverage verifies test files exist per package and meet minimum counts (weight: 0.25).
func checkTestCoverage(root string, opts Options) Category {
	cat := Category{Name: "Test Coverage", Weight: 0.25}

	// Backend test files per package.
	pkgMinTests := map[string]int{
		"api":       40,
		"container": 15,
		"middleware": 10,
		"websocket": 10,
		"config":    5,
		"logging":   5,
	}

	for pkg, minTests := range pkgMinTests {
		pkgDir := filepath.Join(root, "internal", pkg)
		testCount := countFiles(pkgDir, "*_test.go", nil)
		testFuncs := countTestFuncsInDir(pkgDir)

		item := Item{Name: "Tests: internal/" + pkg, MaxScore: 1}
		if testCount > 0 && testFuncs >= minTests {
			item.Status = "pass"
			item.Score = 1
			item.Detail = formatInt(testFuncs) + " test functions (min: " + formatInt(minTests) + ")"
		} else if testCount > 0 {
			item.Status = "partial"
			item.Score = float64(testFuncs) / float64(minTests)
			if item.Score > 1 {
				item.Score = 1
			}
			item.Detail = formatInt(testFuncs) + "/" + formatInt(minTests) + " test functions"
		} else {
			item.Status = "fail"
			item.Detail = "No test files found"
		}
		cat.Items = append(cat.Items, item)
	}

	// Frontend test files.
	feTestFiles := []string{
		"context/WorkspaceContext.test.tsx",
		"services/api.test.ts",
		"services/git.test.ts",
		"services/websocket.test.ts",
		"hooks/useFileLanguage.test.ts",
	}
	for _, tf := range feTestFiles {
		path := filepath.Join(root, opts.FrontendDir, "src", tf)
		cat.Items = append(cat.Items, checkFile(path, "Frontend test: "+tf))
	}

	// Test execution results (if not skipped).
	if !opts.SkipTests {
		goPassed, goFailed, err := runGoTests(root)
		goItem := Item{Name: "Go test suite", MaxScore: 1}
		if err == nil && goFailed == 0 && goPassed > 0 {
			goItem.Status = "pass"
			goItem.Score = 1
			goItem.Detail = formatInt(goPassed) + " tests passing, race detector enabled"
		} else if goPassed > 0 {
			goItem.Status = "partial"
			goItem.Score = float64(goPassed) / float64(goPassed+goFailed)
			goItem.Detail = formatInt(goPassed) + " passing, " + formatInt(goFailed) + " failing"
		} else {
			goItem.Status = "fail"
			goItem.Detail = "Tests failed to run"
		}
		cat.Items = append(cat.Items, goItem)

		fePassed, feFailed, feErr := runFrontendTests(root, opts.FrontendDir)
		feItem := Item{Name: "Frontend test suite", MaxScore: 1}
		if feErr == nil && feFailed == 0 && fePassed > 0 {
			feItem.Status = "pass"
			feItem.Score = 1
			feItem.Detail = formatInt(fePassed) + " tests passing"
		} else if fePassed > 0 {
			feItem.Status = "partial"
			feItem.Score = float64(fePassed) / float64(fePassed+feFailed)
			feItem.Detail = formatInt(fePassed) + " passing, " + formatInt(feFailed) + " failing"
		} else {
			feItem.Status = "fail"
			feItem.Detail = "Tests failed to run"
		}
		cat.Items = append(cat.Items, feItem)
	}

	return cat
}

// checkSpecCompliance verifies that spec requirement IDs are covered by tests/implementation (weight: 0.15).
func checkSpecCompliance(root string, opts Options) Category {
	cat := Category{Name: "Spec Compliance", Weight: 0.15}
	specDir := filepath.Join(root, opts.SpecDir)

	// Extract requirement IDs from spec files.
	specFiles := []string{"api-specification.md", "system-architecture.md", "user-stories.md", "testing-specification.md"}
	allIDs := make(map[string]bool)
	for _, sf := range specFiles {
		for _, id := range extractIDs(filepath.Join(specDir, sf)) {
			allIDs[id] = true
		}
	}

	// Check traceability matrix coverage.
	traceIDs := extractIDs(filepath.Join(specDir, "README.md"))
	item := Item{Name: "Traceability matrix entries", MaxScore: 1}
	if len(traceIDs) >= 5 {
		item.Status = "pass"
		item.Score = 1
		item.Detail = formatInt(len(traceIDs)) + " requirement IDs in traceability matrix"
	} else {
		item.Status = "partial"
		item.Score = float64(len(traceIDs)) / 5.0
		item.Detail = formatInt(len(traceIDs)) + "/5 entries"
	}
	cat.Items = append(cat.Items, item)

	// Check that requirement IDs appear in test files (indicating test coverage of spec).
	testFileIDs := make(map[string]bool)
	testFiles := grepFiles(filepath.Join(root, "internal"), `\[?[A-Z]+-\d+\]?`, "*_test.go")
	for _, tf := range testFiles {
		for _, id := range extractIDs(tf) {
			testFileIDs[id] = true
		}
	}
	// Also check the testing-specification.md which maps tests to requirements.
	for _, id := range extractIDs(filepath.Join(specDir, "testing-specification.md")) {
		testFileIDs[id] = true
	}

	covered := 0
	for id := range allIDs {
		if testFileIDs[id] {
			covered++
		}
	}
	specItem := Item{Name: "Spec IDs with test coverage", MaxScore: 1}
	total := len(allIDs)
	if total > 0 {
		ratio := float64(covered) / float64(total)
		specItem.Score = ratio
		if ratio >= 0.8 {
			specItem.Status = "pass"
		} else if ratio >= 0.5 {
			specItem.Status = "partial"
		} else {
			specItem.Status = "fail"
		}
		specItem.Detail = formatInt(covered) + "/" + formatInt(total) + " requirement IDs covered"
	} else {
		specItem.Status = "fail"
		specItem.Detail = "No requirement IDs found in specs"
	}
	cat.Items = append(cat.Items, specItem)

	// User story acceptance criteria.
	milestones, err := parseMilestones(root, opts.SpecDir)
	if err == nil && len(milestones) > 0 {
		totalCriteria := 0
		completedCriteria := 0
		for _, m := range milestones {
			totalCriteria += m.Criteria
			completedCriteria += m.Completed
		}
		usItem := Item{Name: "User story acceptance criteria", MaxScore: 1}
		if totalCriteria > 0 {
			ratio := float64(completedCriteria) / float64(totalCriteria)
			usItem.Score = ratio
			if ratio >= 0.9 {
				usItem.Status = "pass"
			} else if ratio >= 0.5 {
				usItem.Status = "partial"
			} else {
				usItem.Status = "fail"
			}
			usItem.Detail = formatInt(completedCriteria) + "/" + formatInt(totalCriteria) + " criteria met across " + formatInt(len(milestones)) + " stories"
		}
		cat.Items = append(cat.Items, usItem)
	}

	return cat
}

// checkBuildCI verifies that builds succeed and CI pipeline exists (weight: 0.10).
func checkBuildCI(root string, opts Options) Category {
	cat := Category{Name: "Build & CI", Weight: 0.10}

	if !opts.SkipBuild {
		// Go build.
		goItem := Item{Name: "Go build", MaxScore: 1}
		if err := runGoBuild(root); err == nil {
			goItem.Status = "pass"
			goItem.Score = 1
			goItem.Detail = "go build ./... succeeded"
		} else {
			goItem.Status = "fail"
			goItem.Detail = "go build ./... failed"
		}
		cat.Items = append(cat.Items, goItem)

		// TypeScript check.
		tsItem := Item{Name: "TypeScript check", MaxScore: 1}
		if err := runTypeCheck(root, opts.FrontendDir); err == nil {
			tsItem.Status = "pass"
			tsItem.Score = 1
			tsItem.Detail = "npx tsc --noEmit succeeded"
		} else {
			tsItem.Status = "fail"
			tsItem.Detail = "TypeScript compilation failed"
		}
		cat.Items = append(cat.Items, tsItem)

		// Vite build.
		viteItem := Item{Name: "Frontend production build", MaxScore: 1}
		if err := runViteBuild(root, opts.FrontendDir); err == nil {
			viteItem.Status = "pass"
			viteItem.Score = 1
			viteItem.Detail = "npx vite build succeeded"
		} else {
			viteItem.Status = "fail"
			viteItem.Detail = "Vite build failed"
		}
		cat.Items = append(cat.Items, viteItem)
	}

	// CI pipeline.
	cat.Items = append(cat.Items, checkFile(
		filepath.Join(root, ".github", "workflows", "ci.yml"),
		"CI pipeline: .github/workflows/ci.yml",
	))

	// Check CI has both backend and frontend jobs.
	ciPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	if fileExists(ciPath) {
		hasBackend := len(grepFiles(filepath.Join(root, ".github"), `go build|go test`, "*.yml")) > 0
		hasFrontend := len(grepFiles(filepath.Join(root, ".github"), `npm|vitest|tsc`, "*.yml")) > 0
		ciItem := Item{Name: "CI covers backend + frontend", MaxScore: 1}
		if hasBackend && hasFrontend {
			ciItem.Status = "pass"
			ciItem.Score = 1
			ciItem.Detail = "CI pipeline has Go and Node.js jobs"
		} else {
			ciItem.Status = "partial"
			ciItem.Score = 0.5
			parts := []string{}
			if hasBackend {
				parts = append(parts, "backend")
			}
			if hasFrontend {
				parts = append(parts, "frontend")
			}
			ciItem.Detail = "CI covers: " + strings.Join(parts, ", ")
		}
		cat.Items = append(cat.Items, ciItem)
	}

	return cat
}

// checkDocumentation verifies that docs exist (weight: 0.10).
func checkDocumentation(root string, opts Options) Category {
	cat := Category{Name: "Documentation", Weight: 0.10}

	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "README.md"), "Root README.md"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, ".env.example"), "Environment template: .env.example"))

	// Spec files.
	expectedSpecs := []string{
		"README.md", "api-specification.md", "system-architecture.md",
		"database-schema.md", "user-stories.md", "testing-specification.md",
		"deployment-specification.md",
	}
	for _, sf := range expectedSpecs {
		cat.Items = append(cat.Items, checkFile(
			filepath.Join(root, opts.SpecDir, sf),
			"Spec: "+sf,
		))
	}

	// Go package doc comments.
	pkgs := []string{"api", "config", "container", "middleware", "websocket", "logging"}
	docCount := 0
	for _, pkg := range pkgs {
		pkgDir := filepath.Join(root, "internal", pkg)
		if len(grepFiles(pkgDir, `^// Package\s+\w+`, "*.go")) > 0 {
			docCount++
		}
	}
	docItem := Item{Name: "Go package doc comments", MaxScore: 1}
	docItem.Score = float64(docCount) / float64(len(pkgs))
	if docCount == len(pkgs) {
		docItem.Status = "pass"
		docItem.Detail = formatInt(docCount) + "/" + formatInt(len(pkgs)) + " packages documented"
	} else {
		docItem.Status = "partial"
		docItem.Detail = formatInt(docCount) + "/" + formatInt(len(pkgs)) + " packages documented"
	}
	cat.Items = append(cat.Items, docItem)

	return cat
}

// checkSecurity verifies security measures are in place (weight: 0.10).
func checkSecurity(root string) Category {
	cat := Category{Name: "Security", Weight: 0.10}

	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "internal", "middleware", "auth.go"), "JWT authentication middleware"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "internal", "middleware", "ratelimit.go"), "Rate limiting middleware"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "internal", "middleware", "requestid.go"), "Request correlation middleware"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "internal", "middleware", "cors.go"), "CORS middleware"))

	// Path traversal protection.
	ptFiles := grepFiles(filepath.Join(root, "internal", "api"), `filepath\.Clean`, "*.go")
	ptItem := Item{Name: "Path traversal protection", MaxScore: 1}
	if len(ptFiles) >= 2 {
		ptItem.Status = "pass"
		ptItem.Score = 1
		ptItem.Detail = formatInt(len(ptFiles)) + " files use filepath.Clean for path validation"
	} else {
		ptItem.Status = "fail"
		ptItem.Detail = "Insufficient path traversal protection"
	}
	cat.Items = append(cat.Items, ptItem)

	// Container security.
	secFiles := grepFiles(filepath.Join(root, "internal", "container"), `no-new-privileges`, "*.go")
	secItem := Item{Name: "Container security options", MaxScore: 1}
	if len(secFiles) > 0 {
		secItem.Status = "pass"
		secItem.Score = 1
		secItem.Detail = "no-new-privileges security option configured"
	} else {
		secItem.Status = "fail"
		secItem.Detail = "Missing container security options"
	}
	cat.Items = append(cat.Items, secItem)

	return cat
}

// checkInfrastructure verifies Docker and build infrastructure exists (weight: 0.05).
func checkInfrastructure(root string) Category {
	cat := Category{Name: "Infrastructure", Weight: 0.05}

	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "Dockerfile.api"), "API Dockerfile"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "workspace", "Dockerfile"), "Workspace Dockerfile"))
	cat.Items = append(cat.Items, checkFile(filepath.Join(root, "docker-compose.yml"), "Docker Compose config"))

	// Makefile targets.
	mkPath := filepath.Join(root, "Makefile")
	cat.Items = append(cat.Items, checkFile(mkPath, "Makefile"))

	expectedTargets := []string{"build", "test", "run", "docker-build", "docker-up", "clean"}
	if fileExists(mkPath) {
		foundTargets := 0
		for _, target := range expectedTargets {
			if countPattern(mkPath, `^`+target+`:`) > 0 {
				foundTargets++
			}
		}
		mkItem := Item{Name: "Makefile targets", MaxScore: 1}
		mkItem.Score = float64(foundTargets) / float64(len(expectedTargets))
		if foundTargets == len(expectedTargets) {
			mkItem.Status = "pass"
			mkItem.Detail = formatInt(foundTargets) + "/" + formatInt(len(expectedTargets)) + " required targets present"
		} else {
			mkItem.Status = "partial"
			mkItem.Detail = formatInt(foundTargets) + "/" + formatInt(len(expectedTargets)) + " required targets present"
		}
		cat.Items = append(cat.Items, mkItem)
	}

	return cat
}

// --- Helpers ---

func checkFile(path, name string) Item {
	item := Item{Name: name, MaxScore: 1}
	if fileExists(path) {
		item.Status = "pass"
		item.Score = 1
		item.Detail = "File exists"
	} else {
		item.Status = "fail"
		item.Detail = "File missing: " + path
	}
	return item
}

func checkDir(path, name string) Item {
	item := Item{Name: name, MaxScore: 1}
	if dirExists(path) {
		item.Status = "pass"
		item.Score = 1
		item.Detail = "Directory exists"
	} else {
		item.Status = "fail"
		item.Detail = "Directory missing: " + path
	}
	return item
}

// countTestFuncsInDir counts Go test functions in a single directory.
func countTestFuncsInDir(dir string) int {
	count, _ := countTestFunctions(dir)
	return count
}

func formatInt(n int) string {
	return strconv.Itoa(n)
}
