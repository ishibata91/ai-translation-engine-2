package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	goimportsVersion    = "golang.org/x/tools/cmd/goimports@v0.38.0"
	golangciLintVersion = "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8"
	goCleanarchVersion  = "github.com/roblaszczak/go-cleanarch@v1.2.0"
	govulncheckVersion  = "golang.org/x/vuln/cmd/govulncheck@v1.1.4"
)

func main() {
	if len(os.Args) < 2 {
		exitf("usage: go run ./tools/backendquality <fmt|lint|lint-file|test|check|vuln|watch>")
	}

	var err error
	switch os.Args[1] {
	case "fmt":
		err = runFmt()
	case "lint":
		err = runLint()
	case "lint-file":
		err = runLintFile(os.Args[2:])
	case "test":
		err = runTest()
	case "check":
		err = runFmtCheck()
		if err == nil {
			err = runTest()
		}
	case "vuln":
		err = runVuln()
	case "watch":
		err = runWatch(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		exitf(err.Error())
	}
}

func runFmt() error {
	files, err := goFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	const chunkSize = 64
	for start := 0; start < len(files); start += chunkSize {
		end := min(start+chunkSize, len(files))
		args := append([]string{"run", goimportsVersion, "-w"}, files[start:end]...)
		if err := runCmd(args...); err != nil {
			return err
		}
	}
	return nil
}

func runFmtCheck() error {
	files, err := goFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	const chunkSize = 64
	var changed []string
	for start := 0; start < len(files); start += chunkSize {
		end := min(start+chunkSize, len(files))
		args := append([]string{"run", goimportsVersion, "-l"}, files[start:end]...)
		output, err := runCmdOutput(args...)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				changed = append(changed, line)
			}
		}
	}

	if len(changed) > 0 {
		return fmt.Errorf("goimports required for: %s", strings.Join(changed, ", "))
	}
	return nil
}

func runLint() error {
	patterns, err := lintPatterns()
	if err != nil {
		return err
	}
	args := []string{"run", golangciLintVersion, "run", "--config", ".golangci.yml"}
	args = append(args, patterns...)
	if err := runCmd(args...); err != nil {
		return err
	}
	return runCleanarch()
}

func runLintFile(args []string) error {
	files, err := resolveLintFiles(args)
	if err != nil {
		return err
	}

	if err := runFmtCheckFiles(files); err != nil {
		return err
	}

	patterns := lintPatternsForFiles(files)
	targets := make(map[string]struct{}, len(files))
	for _, file := range files {
		targets[filepath.ToSlash(filepath.Clean(file))] = struct{}{}
	}

	issues, lintErr, err := collectLintIssues(patterns)
	if err != nil {
		return err
	}

	filtered := filterIssuesByFiles(issues, targets)
	if len(filtered) > 0 {
		printLintIssues(filtered)
		return errors.New("backend lint:file failed")
	}

	if lintErr != nil {
		return nil
	}

	return nil
}

func runTest() error {
	return runCmd("test", "./pkg/...")
}

func runVuln() error {
	return runCmd("run", govulncheckVersion, "./pkg/...")
}

func runCleanarch() error {
	args := []string{
		"run",
		goCleanarchVersion,
		"-ignore-tests",
		"-application", "workflow",
		"-interfaces", "controller",
		"-infrastructure", "runtime",
		"-domain", "gateway",
		"./pkg",
	}
	return runCmd(args...)
}

func runWatch(args []string) error {
	cfg, err := parseWatchConfig(args)
	if err != nil {
		return err
	}

	fmt.Printf("watching backend files every %s\n", cfg.interval)
	fmt.Printf("commands: %s\n", strings.Join(cfg.commands, ", "))

	lastSnapshot, err := watchSnapshot()
	if err != nil {
		return err
	}

	if err := runWatchCommands(cfg.commands); err != nil {
		fmt.Fprintf(os.Stderr, "initial run failed: %v\n", err)
	}

	for {
		time.Sleep(cfg.interval)

		current, err := watchSnapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "watch snapshot failed: %v\n", err)
			continue
		}

		if !snapshotsEqual(lastSnapshot, current) {
			fmt.Printf("\nchange detected at %s\n", time.Now().Format(time.RFC3339))
			if err := runWatchCommands(cfg.commands); err != nil {
				fmt.Fprintf(os.Stderr, "watch run failed: %v\n", err)
			}
			lastSnapshot = current
		}
	}
}

func goFiles() ([]string, error) {
	targets := []string{"pkg", "cmd", "tools"}
	files := make([]string, 0, 128)

	for _, name := range []string{"app.go", "main.go"} {
		if _, err := os.Stat(name); err == nil {
			files = append(files, name)
		}
	}

	for _, root := range targets {
		info, err := os.Stat(root)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".go" {
				return nil
			}
			files = append(files, filepath.Clean(path))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	return files, nil
}

func lintPatterns() ([]string, error) {
	patterns := make([]string, 0, 2)

	for _, root := range []string{"pkg", "cmd", "tools"} {
		info, err := os.Stat(root)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			patterns = append(patterns, "./"+filepath.ToSlash(root)+"/...")
		}
	}

	if len(patterns) == 0 {
		return nil, errors.New("no backend package directories found")
	}
	return patterns, nil
}

func lintPatternsForFiles(files []string) []string {
	patterns := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))

	for _, file := range files {
		dir := filepath.Dir(file)
		pattern := "."
		if dir != "." {
			pattern = "./" + filepath.ToSlash(dir)
		}
		if _, ok := seen[pattern]; ok {
			continue
		}
		seen[pattern] = struct{}{}
		patterns = append(patterns, pattern)
	}

	sort.Strings(patterns)
	return patterns
}

func resolveLintFiles(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, errors.New("lint-file requires at least one file or directory")
	}

	files := make([]string, 0, len(args))
	seen := make(map[string]struct{}, len(args))

	for _, target := range args {
		info, err := os.Stat(target)
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("target not found: %s", target)
		}
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			file := filepath.Clean(target)
			if !isBackendLintFile(file) {
				return nil, fmt.Errorf("unsupported backend lint target: %s", target)
			}
			if _, ok := seen[file]; !ok {
				seen[file] = struct{}{}
				files = append(files, file)
			}
			continue
		}

		err = filepath.WalkDir(target, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			file := filepath.Clean(path)
			if !isBackendLintFile(file) {
				return nil
			}
			if _, ok := seen[file]; ok {
				return nil
			}
			seen[file] = struct{}{}
			files = append(files, file)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(files)
	if len(files) == 0 {
		return nil, errors.New("no backend Go files matched the provided targets")
	}
	return files, nil
}

func isBackendLintFile(path string) bool {
	if filepath.Ext(path) != ".go" {
		return false
	}

	cleaned := filepath.ToSlash(filepath.Clean(path))
	if cleaned == "app.go" || cleaned == "main.go" {
		return true
	}

	return strings.HasPrefix(cleaned, "pkg/") || strings.HasPrefix(cleaned, "cmd/") || strings.HasPrefix(cleaned, "tools/")
}

func runFmtCheckFiles(files []string) error {
	const chunkSize = 64
	var changed []string

	for start := 0; start < len(files); start += chunkSize {
		end := min(start+chunkSize, len(files))
		args := append([]string{"run", goimportsVersion, "-l"}, files[start:end]...)
		output, err := runCmdOutput(args...)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				changed = append(changed, line)
			}
		}
	}

	if len(changed) > 0 {
		return fmt.Errorf("goimports required for: %s", strings.Join(changed, ", "))
	}

	return nil
}

type golangciReport struct {
	Issues []golangciIssue `json:"Issues"`
}

type golangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
}

func collectLintIssues(patterns []string) ([]golangciIssue, error, error) {
	args := []string{"run", golangciLintVersion, "run", "--config", ".golangci.yml", "--out-format", "json"}
	args = append(args, patterns...)

	output, runErr := runCmdJSONOutput(args...)
	if output == "" {
		if runErr != nil {
			return nil, runErr, runErr
		}
		return nil, nil, nil
	}

	var report golangciReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		return nil, runErr, fmt.Errorf("failed to parse golangci-lint output: %w", err)
	}

	return report.Issues, runErr, nil
}

func filterIssuesByFiles(issues []golangciIssue, targets map[string]struct{}) []golangciIssue {
	filtered := make([]golangciIssue, 0, len(issues))
	for _, issue := range issues {
		filename := filepath.ToSlash(filepath.Clean(issue.Pos.Filename))
		if _, ok := targets[filename]; ok {
			filtered = append(filtered, issue)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		left := filtered[i]
		right := filtered[j]
		if left.Pos.Filename != right.Pos.Filename {
			return left.Pos.Filename < right.Pos.Filename
		}
		if left.Pos.Line != right.Pos.Line {
			return left.Pos.Line < right.Pos.Line
		}
		if left.Pos.Column != right.Pos.Column {
			return left.Pos.Column < right.Pos.Column
		}
		return left.FromLinter < right.FromLinter
	})

	return filtered
}

func printLintIssues(issues []golangciIssue) {
	for _, issue := range issues {
		fmt.Printf("%s:%d:%d: [%s] %s\n", filepath.ToSlash(issue.Pos.Filename), issue.Pos.Line, issue.Pos.Column, issue.FromLinter, issue.Text)
	}
}

func watchTargets() []string {
	return []string{
		"pkg",
		"cmd",
		"tools/backendquality",
		".golangci.yml",
		"go.mod",
		"go.sum",
		"package.json",
	}
}

type watchConfig struct {
	commands []string
	interval time.Duration
}

func parseWatchConfig(args []string) (watchConfig, error) {
	cfg := watchConfig{
		commands: []string{"check"},
		interval: 2 * time.Second,
	}

	for _, arg := range args {
		switch {
		case arg == "--lint":
			cfg.commands = []string{"check", "lint"}
		case strings.HasPrefix(arg, "--interval="):
			raw := strings.TrimPrefix(arg, "--interval=")
			seconds, err := strconv.Atoi(raw)
			if err != nil || seconds <= 0 {
				return watchConfig{}, fmt.Errorf("invalid interval: %s", raw)
			}
			cfg.interval = time.Duration(seconds) * time.Second
		default:
			return watchConfig{}, fmt.Errorf("unknown watch option: %s", arg)
		}
	}

	return cfg, nil
}

func runWatchCommands(commands []string) error {
	var runErr error
	for _, command := range commands {
		fmt.Printf("==> backend:%s\n", command)
		switch command {
		case "check":
			runErr = runFmtCheck()
			if runErr == nil {
				runErr = runTest()
			}
		case "lint":
			runErr = runLint()
		default:
			runErr = fmt.Errorf("unsupported watch command: %s", command)
		}
		if runErr != nil {
			return runErr
		}
	}
	fmt.Println("watch cycle complete")
	return nil
}

func watchSnapshot() (map[string]time.Time, error) {
	snapshot := make(map[string]time.Time)
	for _, target := range watchTargets() {
		info, err := os.Stat(target)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			snapshot[target] = info.ModTime()
			continue
		}
		err = filepath.WalkDir(target, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if shouldSkipDir(entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			snapshot[filepath.Clean(path)] = info.ModTime()
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return snapshot, nil
}

func snapshotsEqual(a map[string]time.Time, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for path, modTime := range a {
		other, ok := b[path]
		if !ok || !other.Equal(modTime) {
			return false
		}
	}
	return true
}

func shouldSkipDir(name string) bool {
	switch name {
	case "vendor", "node_modules", "build":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

func runCmd(args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "GOFLAGS=")
	}
	return cmd.Run()
}

func runCmdOutput(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	return string(output), err
}

func runCmdJSONOutput(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if runtime.GOOS == "windows" {
		cmd.Env = append(cmd.Env, "GOFLAGS=")
	}
	err := cmd.Run()
	return stdout.String(), err
}

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
