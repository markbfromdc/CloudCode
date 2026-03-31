package evaluate

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// runCommand executes a command in the given directory and returns stdout+stderr and exit code.
func runCommand(dir, name string, args ...string) (string, int, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return string(out), -1, err
		}
	}
	return string(out), exitCode, nil
}

// runGoBuild runs `go build ./...` and returns an error if it fails.
func runGoBuild(root string) error {
	_, code, err := runCommand(root, "go", "build", "./...")
	if err != nil {
		return err
	}
	if code != 0 {
		return &exec.ExitError{}
	}
	return nil
}

// runGoTests runs `go test -race -count=1 ./internal/...` and parses pass/fail counts.
func runGoTests(root string) (passed, failed int, err error) {
	out, _, runErr := runCommand(root, "go", "test", "-v", "-race", "-count=1", "./internal/...")

	passRe := regexp.MustCompile(`--- PASS:`)
	failRe := regexp.MustCompile(`--- FAIL:`)

	for _, line := range strings.Split(out, "\n") {
		if passRe.MatchString(line) {
			passed++
		}
		if failRe.MatchString(line) {
			failed++
		}
	}

	return passed, failed, runErr
}

// runFrontendTests runs `npm test` in the frontend directory and parses pass/fail counts.
func runFrontendTests(root, frontendDir string) (passed, failed int, err error) {
	feDir := filepath.Join(root, frontendDir)
	out, _, runErr := runCommand(feDir, "npx", "vitest", "run")

	// Vitest outputs: "Tests  86 passed (86)" or "Tests  3 failed | 83 passed (86)"
	passRe := regexp.MustCompile(`(\d+)\s+passed`)
	failRe := regexp.MustCompile(`(\d+)\s+failed`)

	if m := passRe.FindStringSubmatch(out); len(m) > 1 {
		passed, _ = strconv.Atoi(m[1])
	}
	if m := failRe.FindStringSubmatch(out); len(m) > 1 {
		failed, _ = strconv.Atoi(m[1])
	}

	return passed, failed, runErr
}

// runTypeCheck runs `npx tsc --noEmit` in the frontend directory.
func runTypeCheck(root, frontendDir string) error {
	feDir := filepath.Join(root, frontendDir)
	_, code, err := runCommand(feDir, "npx", "tsc", "--noEmit")
	if err != nil {
		return err
	}
	if code != 0 {
		return &exec.ExitError{}
	}
	return nil
}

// runViteBuild runs `npx vite build` in the frontend directory.
func runViteBuild(root, frontendDir string) error {
	feDir := filepath.Join(root, frontendDir)
	_, code, err := runCommand(feDir, "npx", "vite", "build")
	if err != nil {
		return err
	}
	if code != 0 {
		return &exec.ExitError{}
	}
	return nil
}
