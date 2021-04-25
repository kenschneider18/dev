<!-- Copyright 2021 Kenneth Schneider

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the License. -->
# Dev

Dev is a tool for managing your development environment inspired by the Golang toolchain of old. Create a directory that will serve as your development path and whenever you wish to clone a git repository or create a new project use this tool to ensure that everything stays organized based on the remote repo's URL.

## Installation

At the moment the only way to install and run this is with the Go toolchain. Assuming you have Go installed on your computer and have added your `$GOPATH` added to your `$PATH`:

`go get github.com/kenschneider18/dev`

#### Configure `$DEVPATH`

- Create a directory on your machine that will serve as the base path for all of the code you've downloaded/installed with this tool.
- Set this as a the `DEVPATH` environment variable in your shell
- Add `$DEVPATH/bin` to your `$PATH` variable in your shell

## Usage

### Get

`get` clones a git repo by it's URL and creates a directory structure matching the repo URL at `$DEVPATH/src`. If the repository has already been cloned this command will do nothing.

```sh
dev get github.com/kenschneider18/rpi-metro-display
```

### Initialize

`init` creates a directory structure matching the passed URL at `$DEVPATH/src`, initializes a git repository, creates a README with the project name, and creates your first commit.

```sh
dev init github.com/mygithubusername/test
```

### Install

Unlike the Go toolchain, there's no easy way to know how to create a binary/install an application. However, for projects that have a makefile with a `devbin` target this command will both clone the code, run `make devbin`, then copy the binary (if there is one) to `$DEVPATH/bin`. If your `$PATH` is properly configured the tool will be available from the command line.

I've created a demo application to demonstrate this:

```sh
dev install github.com/kenschneider18/devtest
```

### Why?

When I first started using Go, I quickly fell in love with the ability to run `go get github.com/<repo>` from any path on my machine, knowing where the repo was cloned, and knowing where I could find it later.

You could accomplish this with a "development" folder and just make sure you change to that directory before cloning, but this becomes unweildy as you accumulate more code. The beauty of the original `go get` command and now this library is that it creates a topological structure based on the repo URL. This structure keeps everything neat and makes repo collisions impossible.

I've personally been using this tool for all of my development for over a year and haven't looked back.
