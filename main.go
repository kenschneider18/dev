package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	devpath := os.Getenv("DEVPATH")
	if devpath == "" {
		log.Fatal("Unset/Invalid DEVPATH")
	}

	// cd to $DEVPATH, crash if the directory doesn't exist
	// they should make the DEVPATH themselves
	err := os.Chdir(devpath)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Invalid DEVPATH %q: %s", devpath, e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open DEVPATH: %s", err.Error())
	}

	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		if err = os.Mkdir(sourceDir, 0777); err != nil {
			log.Fatalf("Failed to create %q directory: %s", sourceDir, err.Error())
		}
	}

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		if err = os.Mkdir(binDir, 0777); err != nil {
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
	if args[0] != "get" && args[0] != "install" {
		log.Fatalf("Unknown command %v", args[0])
	}

	if len(args) != 2 {
		log.Fatalf("Incorrect number of arguments: %q expects 1 argument, found %d", args[0], len(args)-1)
	}

	path := args[1]

	// TODO: add -u flag to prevent this dialogue and update the existing
	// clone ALSO allow passthrough so if someone has run get and then
	// runs install later it will still do the install portion of the code but
	// won't re-clone unless -u is passed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Println("Repo already exists in DEVPATH... Stopping.")
		return
	}

	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		log.Println("Invalid git repo prefix, do not include protocol prefixes such as https:// or git@")
		log.Fatal("Valid repo URL example: github.com/kenschneider18/dev")
	}

	err = os.MkdirAll(path, 0777)
	if err != nil {
		cleanUp(path)
		panic(err)
	}

	err = os.Chdir(path)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Failed to open directory %q: %s", path, e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open directory %q: %s", path, err.Error())
	}

	command := "git"
	commandArgs := []string{"clone", fmt.Sprintf("https://%s", path), "."}

	cmd := exec.Command(command, commandArgs...)
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		cleanUp(path)
		log.Fatalf("Failed to read stderr for git clone: %s", err.Error())
	}

	if err = cmd.Start(); err != nil {
		cleanUp(path)
		log.Fatalf("Failed to clone git repo: %s", err.Error())
	}

	stdErr, _ := ioutil.ReadAll(stdErrPipe)

	if err = cmd.Wait(); err != nil {
		cleanUp(path)
		log.Printf("Failed to clone git repo: %s\n", err.Error())
		log.Fatal(string(stdErr))
	}

	// TODO: only print output of git clone if
	// -v flag passed
	log.Print(string(stdErr))

	if args[0] == "install" {
		install(devpath)
	}
}

func help() {
	fmt.Println("get - clone git repository into organized devpath\ninstall - runs get then builds the project from makefile and installs that to $DEVPATH/bin")
}

func cleanUp(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("Failed to clean up created directory structure: %s\n", err.Error())
	}
}

func install(devpath string) {
	// Run make
	cmd := exec.Command("make", "devbin")
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to read stderr for \"make devbin\": %s", err.Error())
	}

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to read stdout for \"make devbin\": %s", err.Error())
	}

	if err = cmd.Start(); err != nil {
		log.Fatalf("Failed to make: %s", err.Error())
	}

	stdErr, _ := ioutil.ReadAll(stdErrPipe)
	stdOut, _ := ioutil.ReadAll(stdOutPipe)

	if err = cmd.Wait(); err != nil {
		log.Printf("Failed to make: %s\n", err.Error())
		log.Fatal(string(stdErr))
	}

	// TODO: only print output of make
	// -v flag passed
	log.Printf(string(stdOut))

	// open devbin directory
	err = os.Chdir(devbinDir)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Failed to open directory %q: %s", devbinDir, e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open directory %q: %s", devbinDir, err.Error())
	}

	filesToMove := make([]string, 0, 5)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Failed to list files in \"devbin\"")
			return err
		}

		if info.IsDir() || path == "." || path == ".." {
			return nil
		}

		filesToMove = append(filesToMove, path)

		return nil
	})

	log.Println("Installing binaries:", filesToMove)

	for _, file := range filesToMove {
		err = os.Rename(file, fmt.Sprintf("%s/%s/%s", devpath, binDir, file))
		if err != nil {
			log.Fatalf("Failed to move file: %s", file)
		}
	}

	// open parent directory
	err = os.Chdir(prevDir)
	if e, ok := err.(*os.PathError); ok && err != nil {
		log.Fatalf("Failed to open parent directory: %s", e.Unwrap().Error())
	} else if err != nil {
		log.Fatalf("Failed to open parent directory: %s", err.Error())
	}

	err = os.RemoveAll(devbinDir)
	if err != nil {
		log.Fatalf("Failed to remove directory %q: %s", devbinDir, err.Error())
	}
}
