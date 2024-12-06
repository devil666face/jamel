.DEFAULT_GOAL := help

BIN = $(shell pwd)/bin
include build.env
export $(shell sed 's/=.*//' build.env)
$(shell [ -f bin ] || mkdir -p $(BIN))

# SBOM = sbom
CVE = cve
GOBIN = go
PATH := $(BIN):$(PATH)
GOARCH = amd64
LDFLAGS = -extldflags '-static' -w -s -buildid= 
GCFLAGS = all=-trimpath=$(shell pwd) -dwarf=false -l
ASMFLAGS = all=-trimpath=$(shell pwd)

help:
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.crop:
	for file in $(wildcard $(BIN)/*); do \
		strip $$file; \
		objcopy --strip-unneeded $$file; \
	done

build: build-cve .crop ## build all

# build-sbom: ## build sbom
# 	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
# 		$(GOBIN) build -ldflags="$(LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
# 		-o $(BIN)/$(SBOM) cmd/$(SBOM)/main.go

build-cve: ## build cves
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
		$(GOBIN) build -ldflags="$(LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
		-o $(BIN)/$(CVE) cmd/$(CVE)/main.go
