NAME=ddls2m
VERSION=0.0.1
RELEASE_DIR=$(CURDIR)/release

deps:
	go mod download
	go mod tidy

clean:
	rm -rf $(RELEASE_DIR)/*

build:
	go build -o $(RELEASE_DIR)/$(NAME)_$(GOOS)_$(GOARCH) cmd/$(NAME)/main.go

all: clean build-linux-amd64 build-darwin-amd64
build-linux-amd64:
	@$(MAKE) build GOOS=linux GOARCH=amd64
build-darwin-amd64:
	@$(MAKE) build GOOS=darwin GOARCH=amd64
.PHONY: release
release:
	ghr -u nakatamixi -t $(shell cat github_token) -replace $(VERSION) $(RELEASE_DIR)
