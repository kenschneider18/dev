package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kenschneider18/dev/pkg/executor"
)

const (
	sourceDir = "src"
	binDir    = "bin"
	devbinDir = "devbin"
	prevDir   = ".."
)

func main() {
	// command line args minus program name
	args := os.Args[1:]

	if len(args) <= 0 || args[0] == "help" {
		help()
		return
	}

	// look for DEVPATH as place to put all dev dependencies
	devPath := os.Getenv("DEVPATH")
	if devPath == "" {
		log.Fatal("Unset/Invalid DEVPATH environment variable")
	}

	// cd to $DEVPATH, crash if the directory doesn't exist
	// they should make the DEVPATH themselves
	err := os.Chdir(devPath)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Invalid DEVPATH %q: %s", devPath, e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open DEVPATH: %s", err.Error())
	}

	// TODO: figure out if these permissions settings make sense
	// or if they should be changed
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		if err = os.Mkdir(sourceDir, 0755); err != nil {
			log.Fatalf("Failed to create %q directory: %s", sourceDir, err.Error())
		}
	}

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		if err = os.Mkdir(binDir, 0755); err != nil {
			log.Fatalf("Failed to create %q directory: %s", binDir, err.Error())
		}
	}

	err = os.Chdir(sourceDir)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Failed to open %q: %s", sourceDir, e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open %q: %s", sourceDir, err.Error())
	}

	// parse command
	cmd, ok := executor.ValidCommands[args[0]]
	if !ok {
		log.Fatalln("Unknown command")
	}

	executor := executor.New(devbinDir, binDir, sourceDir, prevDir, devPath, cmd)
	err = executor.Execute(args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: make this a more expansive help section
func help() {
	fmt.Println("get - clone git repository into organized devpath\ninstall - runs get then builds the project from makefile and installs that to $DEVPATH/bin")
}
