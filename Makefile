GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOARM ?= $(shell go env GOARM)

.PHONY: all
all: a2s

.PHONY: a2s
a2s:
	go build ./cmd/a2s

.PHONY: dist
dist: a2s-$(GOOS)-$(GOARCH).tar.gz

.PHONY: dist-all
dist-all: a2s-linux-amd64.tar.gz
dist-all: a2s-linux-arm64.tar.gz
dist-all: a2s-windows-amd64.zip

a2s-linux-amd64.tar.gz: GOOS := linux
a2s-linux-amd64.tar.gz: GOARCH := amd64
a2s-linux-arm64.tar.gz: GOOS := linux
a2s-linux-arm64.tar.gz: GOARCH := arm64
a2s-windows-amd64.zip: GOOS := windows
a2s-windows-amd64.zip: GOARCH := amd64
a2s-windows-amd64.zip: EXT := .exe
a2s-%.tar.gz:
	mkdir -p "dist/a2s-$*"
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o "dist/a2s-$*/a2s$(EXT)" -ldflags "-w -s" ./cmd/a2s
	cp a2s.conf "dist/a2s-$*"
	cd dist && tar czf "$@" "a2s-$*"

a2s-%.zip:
	mkdir -p "dist/a2s-$*"
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o "dist/a2s-$*/a2s$(EXT)" -ldflags "-w -s" ./cmd/a2s
	cp a2s.conf "dist/a2s-$*"
	cd dist && zip -r "$@" "a2s-$*"

.PHONY: clean
clean:
	rm -f a2s{,.exe}
	rm -rf dist
