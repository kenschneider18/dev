VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -X main.version=$(VERSION)

build: main.go
	go build -ldflags "$(LDFLAGS)" -o dev main.go

devbin: main.go
	mkdir -p devbin
	go build -ldflags "$(LDFLAGS)" -o devbin/dev main.go
