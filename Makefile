.PHONY: build-agent build-server

# .SILENT:

build-agent:
	cd ./cmd/agent && \
	go build -ldflags "-X main.buildVersion=$(BuildVersion) \
		-X 'main.buildDate=$(BuildDate)' \
		-X 'main.buildCommit=$(BuildCommit)'"

build-server:
	cd ./cmd/server && \
	go build -ldflags "-X main.buildVersion=$(BuildVersion) \
		-X 'main.buildDate=$(BuildDate)' \
		-X 'main.buildCommit=$(BuildCommit)'"

# BuildVersion can be provided like this: `make build-agent BuildVersion=v1.2.3`
BuildVersion := v0.1.0
BuildDate := $(shell date +'%Y/%m/%d %H:%M:%S')
BuildCommit := $(shell git rev-parse HEAD)
