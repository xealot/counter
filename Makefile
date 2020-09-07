.DEFAULT_GOAL := build

build: amd64 arm64 darwin
amd64:
	env GOOS=linux GOARCH=amd64 go build -o bin/counter_amd64 counter.go
arm64:
	env GOOS=linux GOARCH=arm64 go build -o bin/counter_arm64 counter.go
darwin:
	env GOOS=darwin GOARCH=amd64 go build -o bin/counter_darwin counter.go