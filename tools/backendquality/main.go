package main

import (
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
	govulncheckVersion  = "golang.org/x/vuln/cmd/govulncheck@v1.1.4"
)

func main() {
	if len(os.Args) < 2 {
		exitf("usage: go run ./tools/backendquality <fmt|lint|test|check|vuln|watch>")
	}

	var err error
	switch os.Args[1] {
	case "fmt":
		err = runFmt()
	case "lint":
		err = runLint()
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
	return runCmd(args...)
}

func runTest() error {
	return runCmd("test", "./pkg/...")
}

func runVuln() error {
	return runCmd("run", govulncheckVersion, "./pkg/...")
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
	targets := []string{"pkg", "cmd"}
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

	for _, root := range []string{"pkg", "cmd"} {
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

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
