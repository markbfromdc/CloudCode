package evaluate

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists returns true if the path exists and is a directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// countFiles counts files matching a glob pattern under root, excluding specified directories.
func countFiles(root, pattern string, excludeDirs []string) int {
	count := 0
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			for _, ex := range excludeDirs {
				if name == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}
		matched, _ := filepath.Match(pattern, d.Name())
		if matched {
			count++
		}
		return nil
	})
	return count
}

// countLOC counts non-empty lines of code for files with given extensions, excluding specified directories.
func countLOC(root string, extensions, excludeDirs []string) int {
	total := 0
	extSet := make(map[string]bool)
	for _, ext := range extensions {
		extSet[ext] = true
	}

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			for _, ex := range excludeDirs {
				if d.Name() == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}
		ext := filepath.Ext(d.Name())
		if !extSet[ext] {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) != "" {
				total++
			}
		}
		return nil
	})
	return total
}

// countTestFunctions counts Go test functions (func Test*) in _test.go files under root.
func countTestFunctions(root string) (int, error) {
	testFuncRe := regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)
	count := 0

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if testFuncRe.MatchString(scanner.Text()) {
				count++
			}
		}
		return nil
	})
	return count, err
}

// countFrontendTestFunctions counts test cases (it( or test( calls) in frontend test files.
func countFrontendTestFunctions(root, frontendDir string) int {
	testCallRe := regexp.MustCompile(`\b(it|test)\s*\(`)
	count := 0
	feRoot := filepath.Join(root, frontendDir, "src")

	filepath.WalkDir(feRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".test.ts") && !strings.HasSuffix(d.Name(), ".test.tsx") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if testCallRe.MatchString(scanner.Text()) {
				count++
			}
		}
		return nil
	})
	return count
}

// grepFiles returns file paths under root where any line matches the regex pattern.
// Only files matching the glob are searched. Excludes node_modules and vendor.
func grepFiles(root, pattern, glob string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	var matches []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == "vendor" || d.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if glob != "" {
			matched, _ := filepath.Match(glob, d.Name())
			if !matched {
				return nil
			}
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if re.MatchString(scanner.Text()) {
				matches = append(matches, path)
				return nil // Found in this file, move to next.
			}
		}
		return nil
	})
	return matches
}

// countPattern counts the number of lines matching a regex in a single file.
func countPattern(filePath, pattern string) int {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0
	}
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if re.MatchString(scanner.Text()) {
			count++
		}
	}
	return count
}

// extractIDs extracts all requirement IDs matching [AREA-NNN] or bare AREA-NNN patterns from a file.
// Matches both `[API-001]` and `API-001` (e.g., in markdown tables).
func extractIDs(filePath string) []string {
	// Match bracketed [API-001] or bare API-001 (word boundary).
	re := regexp.MustCompile(`\b([A-Z]+-\d+)\b`)
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	seen := make(map[string]bool)
	var ids []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		for _, match := range re.FindAllStringSubmatch(scanner.Text(), -1) {
			id := match[1]
			// Filter out common false positives (e.g., SHA-256, UTF-8).
			if id == "SHA-256" || id == "UTF-8" || id == "HS-256" || strings.HasPrefix(id, "ES-") {
				continue
			}
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}
