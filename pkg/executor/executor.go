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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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

// Making a commit with codespaces
var (
	// ValidCommands maps a command line argument to a command
	ValidCommands = map[string]Command{
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
		devbinDir   string
		binDir      string
		srcDir      string
		prevDir     string
		devPath     string
		command     Command
	}
)

// New returns a new executor to run a command with
func New(devbinDir, binDir, srcDir, prevDir, devPath string, command Command) *Executor {
	return &Executor{
		devbinDir: devbinDir,
		binDir:    binDir,
		srcDir:    srcDir,
		prevDir:   prevDir,
		devPath:   devPath,
		command:   command,
	}
}

func (e *Executor) Execute(args []string) error {
	// this should probably turn into a map at some point
	switch e.command {
	case GET:
		if len(args) != 1 {
			return errors.Errorf("get expects 1 argument, found %d", len(args))
		}
		return e.clone(args[0])
	case INSTALL:
		if len(args) != 1 {
			return errors.Errorf("install expects 1 argument, found %d", len(args))
		}
		err := e.clone(args[0])
		if err != nil {
			return err
		}
		return e.install()
	case INITIALIZE:
		var path string
		var language string
		if len(args) < 1 || len(args) > 2 {
			return errors.Errorf("init expects at most 2 arguments, found %d", len(args))
		} else if len(args) == 1 {
			path = args[0]
		} else {
			path = args[0]
			language = args[1]
		}

		return e.init(path, language)
	default:
		return errors.Errorf("command not found")
	}
}

func (e *Executor) clone(path string) error {
	// TODO: allow this to be done with or without a prefix
	// default clones via HTTPS but supports either protocol

	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		return errors.New("invalid git repo prefix: do not include protocol prefixes such as https:// or git@")
	}

	// TODO: add -u flag to prevent this dialogue and update the existing
	// clone ALSO allow passthrough so if someone has run get and then
	// runs install later it will still do the install portion of the code but
	// won't re-clone unless -u is passed
	err := e.makeDir(path, false)
	if err != nil {
		return err
	}

	command := "git"
	commandArgs := []string{"clone", fmt.Sprintf("https://%s", path), "."}

	cmd := exec.Command(command, commandArgs...)
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		e.cleanUp(path)
		return errors.Wrap(err, "failed to read from git")
	}

	if err = cmd.Start(); err != nil {
		e.cleanUp(path)
		return errors.Wrap(err, "failed to clone git repo")
	}

	stdErr, _ := ioutil.ReadAll(stdErrPipe)

	if err = cmd.Wait(); err != nil {
		e.cleanUp(path)
		return errors.Wrapf(err, "failed to clone repo: %s", string(stdErr))
	}

	// TODO: only print output of git clone if
	// -v flag passed
	log.Print(string(stdErr))

	return nil
}

func (e *Executor) install() error {
	// Run make
	cmd := exec.Command("make", "devbin")
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "failed to read output from make")
	}

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to read output from make")
	}

	if err = cmd.Start(); err != nil {
		return errors.Wrap(err, "failed running make devbin")
	}

	stdErr, _ := ioutil.ReadAll(stdErrPipe)
	stdOut, _ := ioutil.ReadAll(stdOutPipe)

	if err = cmd.Wait(); err != nil {
		return errors.Wrapf(err, "failed to run make devbin %s", string(stdErr))
	}

	// TODO: only print output of make
	// -v flag passed
	log.Print(string(stdOut))

	// open devbin directory
	err = os.Chdir(e.devbinDir)
	if pathErr, ok := err.(*os.PathError); ok && err != nil {
		return errors.Wrapf(pathErr.Unwrap(), "failed to open %s", e.devbinDir)
	} else if err != nil {
		return errors.Wrapf(err, "failed to open %s", e.devbinDir)
	}

	filesToMove := make([]string, 0, 5)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to ls devbin")
		}

		if info.IsDir() || path == "." || path == ".." {
			return nil
		}

		filesToMove = append(filesToMove, path)

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to walk directory")
	}

	log.Println("Installing binaries:", filesToMove)

	for _, file := range filesToMove {
		err = os.Rename(file, fmt.Sprintf("%s/%s/%s", e.devPath, e.binDir, file))
		if err != nil {
			log.Fatalf("Failed to move file: %s", file)
		}
	}

	// open parent directory
	err = os.Chdir(e.prevDir)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Failed to open parent directory: %s", e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open parent directory: %s", err.Error())
	}

	err = os.RemoveAll(e.devbinDir)
	if err != nil {
		log.Fatalf("Failed to remove directory %q: %s", e.devbinDir, err.Error())
	}
	return nil
}

func (e *Executor) init(path, language string) error {
	err := e.makeDir(path, false)
	if err != nil {
		return err
	}

	output, err := runCommand("git", []string{"init"})
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
	err = ioutil.WriteFile("README.md", readme, 0544)
	if err != nil {
		return err
	}

	if strings.EqualFold(language, "go") {
		output, err := runCommand("go", []string{"mod", "init", path})
		if err != nil {
			return err
		}

		// TODO: only print output if
		// -v flag passed
		if len(output) > 0 {
			log.Print(output)
		}
	}

	output, err = runCommand("git", []string{"add", "."})
	if err != nil {
		return err
	}

	// TODO: only print output if
	// -v flag passed
	if len(output) > 0 {
		log.Print(output)
	}

	output, err = runCommand("git", []string{"commit", "-m", `"Initialize repository"`})
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

	err = os.Chdir(path)
	if pe, ok := err.(*os.PathError); ok && err != nil {
		e.cleanUp(path)
		return errors.Wrapf(pe.Unwrap(), "failed to open directory %q", path)
	} else if err != nil {
		e.cleanUp(path)
		return errors.Wrapf(err, "failed to open directory %q", path)
	}

	return nil
}

func (e *Executor) cleanUp(path string) error {
	err := os.RemoveAll(fmt.Sprintf("%s/%s/%s", e.devPath, e.srcDir, path))
	if err != nil {
		return errors.Wrap(err, "failed to clean up created directory structure")
	}
	return nil
}

func runCommand(command string, args []string) (string, error) {
	cmd := exec.Command(command, args...)
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.Wrap(err, "failed to start command")
	}

	if err = cmd.Start(); err != nil {
		return "", errors.Wrap(err, "failed to run command")
	}

	stdErr, _ := ioutil.ReadAll(stdErrPipe)

	if err = cmd.Wait(); err != nil {
		return "", errors.Wrapf(err, "failed to run command with error: %s", string(stdErr))
	}

	return string(stdErr), nil
}
