build: main.go
		go build -o dev main.go

devbin: main.go
		mkdir -p devbin
		go build -o devbin/dev main.go
