// Package executor runs defined commands
/* Copyright 2021 Kenneth Schneider

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the License. */
package executor

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// GET downloads a repo creating a directory structure
	// based on the repo URL
	GET = iota
	// INSTALL runs GET then takes the extra step of running
	// `make devbin` and copies the resulting binary to $DEVPATH/bin
	// If $DEVPATH/bin is part of your $PATH, these binaries will be
	// immediately accesible via the command line
	INSTALL
	// INITIALIZE creates an empty directory at the specified repo URL
	// within $DEVPATH. The programming language of choice can be passed
	// for supported languages to do additional setup.
	INITIALIZE
)

var (
	// ValidCommands maps a command line argument to a command
	validCommands = map[string]Command{
		"get":     GET,
		"install": INSTALL,
		"init":    INITIALIZE,
	}
)

type (
	// Command represents the action to be executed
	Command int

	// Executor runs a command
	Executor struct {
		devbinDir string
		binDir    string
		srcDir    string
		devPath   string
		workDir   string
		command   Command
	}
)

// New returns a new executor to run a command with
func New(devbinDir, binDir, srcDir, devPath, command string) (*Executor, error) {
	cmd, ok := validCommands[command]
	if !ok {
		return nil, fmt.Errorf("invalid command %q", command)
	}

	return &Executor{
		devbinDir: devbinDir,
		binDir:    binDir,
		srcDir:    srcDir,
		devPath:   devPath,
		workDir:   filepath.Join(devPath, srcDir),
		command:   cmd,
	}, nil
}

func (e *Executor) Execute(args []string) error {
	// this should probably turn into a map at some point
	switch e.command {
	case GET:
		if len(args) != 1 {
			return fmt.Errorf("get expects 1 argument, found %d", len(args))
		}
		_, err := e.clone(args[0])
		return err
	case INSTALL:
		if len(args) != 1 {
			return fmt.Errorf("install expects 1 argument, found %d", len(args))
		}
		repoDir, err := e.clone(args[0])
		if err != nil {
			return err
		}
		return e.install(repoDir)
	case INITIALIZE:
		var path string
		var language string
		if len(args) < 1 || len(args) > 2 {
			return fmt.Errorf("init expects at most 2 arguments, found %d", len(args))
		} else if len(args) == 1 {
			path = args[0]
		} else {
			path = args[0]
			language = args[1]
		}

		return e.init(path, language)
	default:
		return fmt.Errorf("command not found")
	}
}

func (e *Executor) clone(path string) (string, error) {
	repoPath, cloneURL, err := normalizeClonePath(path)
	if err != nil {
		return "", err
	}

	absRepoPath := filepath.Join(e.workDir, repoPath)

	// TODO: add -u flag to prevent this dialogue and update the existing
	// clone ALSO allow passthrough so if someone has run get and then
	// runs install later it will still do the install portion of the code but
	// won't re-clone unless -u is passed
	err = e.makeDir(absRepoPath, false)
	if err != nil {
		return "", err
	}

	output, err := runCommand(absRepoPath, "git", []string{"clone", cloneURL, "."})
	if err != nil {
		e.cleanUp(absRepoPath)
		return "", fmt.Errorf("failed to clone repo: %w", err)
	}

	// TODO: only print output of git clone if
	// -v flag passed
	log.Print(output)

	return absRepoPath, nil
}

func normalizeClonePath(raw string) (string, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", errors.New("repo path is required")
	}

	if strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "http://") {
		repoPath, cloneURL, err := parseHTTPRepoURL(trimmed)
		if err != nil {
			return "", "", err
		}
		return repoPath, cloneURL, nil
	}

	if strings.HasPrefix(trimmed, "git@") {
		repoPath, err := parseGitSSHPath(trimmed)
		if err != nil {
			return "", "", err
		}
		return repoPath, trimmed, nil
	}

	repoPath, err := normalizeRepoPath(trimmed)
	if err != nil {
		return "", "", err
	}
	return repoPath, fmt.Sprintf("https://%s", repoPath), nil
}

func parseHTTPRepoURL(raw string) (string, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("invalid http(s) repo URL: %w", err)
	}

	if (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return "", "", errors.New("invalid http(s) repo URL")
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	if path == "" {
		return "", "", errors.New("invalid http(s) repo URL: missing repository path")
	}

	repoPath, err := normalizeRepoPath(fmt.Sprintf("%s/%s", parsed.Host, path))
	if err != nil {
		return "", "", err
	}

	// Warn if credentials are embedded in the URL — they will be visible in
	// process listings while git clone is running.
	if parsed.User != nil {
		log.Println("warning: embedded credentials detected in repo URL; these will be visible in process listings")
	}
	return repoPath, raw, nil
}

func parseGitSSHPath(raw string) (string, error) {
	withoutPrefix := strings.TrimPrefix(raw, "git@")
	if withoutPrefix == "" {
		return "", errors.New("invalid ssh repo URL")
	}

	// Standard SSH git URLs use colon as the separator between host and path
	// (e.g. git@github.com:org/repo). Slash is accepted as a fallback for
	// non-standard formats (e.g. git@github.com/org/repo).
	separator := ":"
	if !strings.Contains(withoutPrefix, separator) {
		separator = "/"
	}

	parts := strings.SplitN(withoutPrefix, separator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", errors.New("invalid ssh repo URL")
	}

	return normalizeRepoPath(fmt.Sprintf("%s/%s", parts[0], parts[1]))
}

func normalizeRepoPath(path string) (string, error) {
	normalized := strings.Trim(path, "/")
	normalized = strings.TrimSuffix(normalized, ".git")
	normalized = strings.TrimSuffix(normalized, "/")
	// Require at least host/org/repo (2 slashes minimum); deeper paths like
	// gitlab.com/group/subgroup/project are also valid.
	if normalized == "" || strings.Count(normalized, "/") < 2 {
		return "", errors.New("invalid repo path: requires at least host/org/repo")
	}
	return normalized, nil
}

func (e *Executor) install(repoDir string) error {
	// Run make devbin in the repo directory
	cmd := exec.Command("make", "devbin")
	cmd.Dir = repoDir
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to read output from make: %w", err)
	}

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to read output from make: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed running make devbin: %w", err)
	}

	stdErr, _ := io.ReadAll(stdErrPipe)
	stdOut, _ := io.ReadAll(stdOutPipe)
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("failed to run make devbin %s: %w", string(stdErr), err)
	}

	// TODO: only print output of make
	// -v flag passed
	log.Printf("%s", stdOut)

	devbinDir := filepath.Join(repoDir, e.devbinDir)
	filesToMove := make([]string, 0, 5)
	err = filepath.Walk(devbinDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to ls devbin: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		filesToMove = append(filesToMove, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	log.Println("Installing binaries: ", filesToMove)

	for _, file := range filesToMove {
		dst := filepath.Join(e.devPath, e.binDir, filepath.Base(file))
		err = os.Rename(file, dst)
		if err != nil {
			return fmt.Errorf("failed to move file %q to %q: %w", file, dst, err)
		}
	}

	err = os.RemoveAll(devbinDir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %q: %w", devbinDir, err)
	}
	return nil
}

func (e *Executor) init(path, language string) error {
	absPath := filepath.Join(e.workDir, path)
	err := e.makeDir(absPath, false)
	if err != nil {
		return err
	}

	output, err := runCommand(absPath, "git", []string{"init"})
	if err != nil {
		return err
	}

	// TODO: only print output if
	// -v flag passed
	if len(output) > 0 {
		log.Print(output)
	}

	splitPath := strings.Split(path, "/")
	readme := []byte(fmt.Sprintf("# %s\n", splitPath[len(splitPath)-1]))
	err = os.WriteFile(filepath.Join(absPath, "README.md"), readme, 0644)
	if err != nil {
		return err
	}

	if strings.EqualFold(language, "go") {
		output, err := runCommand(absPath, "go", []string{"mod", "init", path})
		if err != nil {
			return err
		}

		// TODO: only print output if
		// -v flag passed
		if len(output) > 0 {
			log.Print(output)
		}
	}

	output, err = runCommand(absPath, "git", []string{"add", "."})
	if err != nil {
		return err
	}

	// TODO: only print output if
	// -v flag passed
	if len(output) > 0 {
		log.Print(output)
	}

	output, err = runCommand(absPath, "git", []string{"commit", "-m", "Initialize repository"})
	if err != nil {
		return err
	}

	// TODO: only print output if
	// -v flag passed
	if len(output) > 0 {
		log.Print(output)
	}

	return nil
}

func (e *Executor) makeDir(path string, ignoreAlreadyExists bool) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) && !ignoreAlreadyExists {
		return errors.New("repo already exists in DEVPATH")
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		e.cleanUp(path)
		return err
	}

	return nil
}

func (e *Executor) cleanUp(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to clean up created directory structure: %w", err)
	}
	return nil
}

func runCommand(dir, command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	if err = cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	stdErr, _ := io.ReadAll(stdErrPipe)

	if err = cmd.Wait(); err != nil {
		return "", fmt.Errorf("failed to run command with error: %s: %w", string(stdErr), err)
	}

	return string(stdErr), nil
}
