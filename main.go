// Package main is the main package for dev
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
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kenschneider18/dev/pkg/executor"
)

const (
	sourceDir = "src"
	binDir    = "bin"
	devbinDir = "devbin"
)

var version = "dev"

func main() {
	// command line args minus program name
	args := os.Args[1:]

	// strip flags before looking at the command
	verbose := false
	update := false
	filtered := args[:0]
	for _, a := range args {
		switch a {
		case "-v", "--verbose":
			verbose = true
		case "-u", "--update":
			update = true
		default:
			filtered = append(filtered, a)
		}
	}
	args = filtered

	if len(args) == 0 || args[0] == "help" {
		help()
		return
	}

	if args[0] == "version" || args[0] == "--version" {
		fmt.Printf("dev %s\n", version)
		return
	}

	// look for DEVPATH as place to put all dev dependencies
	devPath := os.Getenv("DEVPATH")
	if devPath == "" {
		log.Fatal("Unset/Invalid DEVPATH environment variable")
	}

	// verify $DEVPATH exists and is a directory — users must create it themselves
	info, err := os.Stat(devPath)
	if os.IsNotExist(err) {
		log.Fatalf("Invalid DEVPATH %q: directory does not exist", devPath)
	} else if err != nil {
		log.Fatalf("Failed to open DEVPATH: %s", err.Error())
	} else if !info.IsDir() {
		log.Fatalf("Invalid DEVPATH %q: not a directory", devPath)
	}

	// TODO: figure out if these permissions settings make sense
	// or if they should be changed
	if _, err := os.Stat(filepath.Join(devPath, sourceDir)); os.IsNotExist(err) {
		if err = os.Mkdir(filepath.Join(devPath, sourceDir), 0755); err != nil {
			log.Fatalf("Failed to create %q directory: %s", sourceDir, err.Error())
		}
	}

	if _, err := os.Stat(filepath.Join(devPath, binDir)); os.IsNotExist(err) {
		if err = os.Mkdir(filepath.Join(devPath, binDir), 0755); err != nil {
			log.Fatalf("Failed to create %q directory: %s", binDir, err.Error())
		}
	}

	executor, err := executor.New(devbinDir, binDir, sourceDir, devPath, args[0], verbose, update)
	if err != nil {
		log.Fatalln(err)
	}

	err = executor.Execute(args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

func help() {
	fmt.Printf(`dev - manage your development environment

Usage:
  dev [-v] <command> [arguments]

Commands:
  get <repo>            Clone a repository to $DEVPATH/src/<host>/<org>/<repo>
  init <repo> [lang]    Create a new repository in $DEVPATH/src
  install <repo>        Clone a repository and install its binary to $DEVPATH/bin
  version               Show version information
  help                  Show this help message

Flags:
  -v, --verbose         Print output from git and make subcommands
  -u, --update          Pull latest changes if repo already exists (get/install)

Environment:
  DEVPATH               Base directory for all repositories and binaries

Version: %s
`, version)
}
