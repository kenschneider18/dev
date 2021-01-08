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

Dev is a tool for managing your development environment inspired by the Golang toolchain of old.

## What can I do with this?

### get

`get` creates a topological 

```sh
dev
```

### Why?

When I first started using Go, I quickly fell in love with the ability to run `go get github.com/<repo>` from any path on my machine, knowing where the repo was cloned, and knowing where I could find it later.

You could accomplish this with a "development" folder and just make sure you change to that directory before cloning, but this becomes unweildy as you accumulate more code. The beauty of the original `go get` command and now this library is that it creates a topological structure based on the repo URL. This structure keeps everything neat and makes repo collisions impossible.