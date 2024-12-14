.DEFAULT_GOAL := help

BIN = $(shell pwd)/bin
include build.env
export $(shell sed 's/=.*//' build.env)
$(shell [ -f bin ] || mkdir -p $(BIN))

APP = jamel
SBOM = sbom
CLIENT = client
SERVER = server
ADMIN = admin
GOBIN = go
PATH := $(BIN):$(PATH)
GOARCH = amd64
LDFLAGS = -extldflags '-static' -w -s -buildid= 
GCFLAGS = all=-trimpath=$(shell pwd) -dwarf=false -l
ASMFLAGS = all=-trimpath=$(shell pwd)
STRIP_FILES = $(BIN)/$(APP)-$(CLIENT) $(BIN)/$(APP)-$(SERVER) $(BIN)/$(APP)-$(ADMIN)_linux $(BIN)/$(APP)-$(ADMIN)_windows.exe

help:
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.crop:
	for file in $(STRIP_FILES); do \
		strip $$file; \
		objcopy --strip-unneeded $$file; \
	done

build: build-client build-server build-admin-linux .crop ## build
release: build-client build-server release-admin .crop ## release
release-admin: build-admin-linux build-admin-windows build-admin-darwin-amd64 build-admin-darwin-arm64 ## build all admin bin

build-sbom: ## build sbom
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
		$(GOBIN) build -ldflags="$(LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
		-o $(BIN)/$(SBOM) cmd/$(SBOM)/main.go

build-client: ## build client
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
		$(GOBIN) build -ldflags="$(LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
		-o $(BIN)/$(APP)-$(CLIENT) cmd/$(CLIENT)/main.go

build-server: ## build server
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
		$(GOBIN) build -ldflags="$(LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
		-o $(BIN)/$(APP)-$(SERVER) cmd/$(SERVER)/main.go

build-admin-linux: ## build admin for linux
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) \
	  $(GOBIN) build -ldflags="$(LINUX_LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
	  -o $(BIN)/$(APP)-$(ADMIN)_linux cmd/$(ADMIN)/main.go

build-admin-windows: ## build admin for windows
	CGO_ENABLED=0 GOOS=windows GOARCH=$(GOARCH) \
	  $(GOBIN) build -ldflags="$(WINDOWS_LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
	  -o $(BIN)/$(APP)-$(ADMIN)_windows.exe cmd/$(ADMIN)/main.go

build-admin-darwin-amd64: ## build admin for darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=$(GOARCH) \
	  $(GOBIN) build -ldflags="$(LINUX_LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
	  -o $(BIN)/$(APP)-$(ADMIN)_darwin_amd64 cmd/$(ADMIN)/main.go

build-admin-darwin-arm64: ## build admin for darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
	  $(GOBIN) build -ldflags="$(LINUX_LDFLAGS)" -trimpath -gcflags="$(GCFLAGS)" -asmflags="$(ASMFLAGS)" \
	  -o $(BIN)/$(APP)-$(ADMIN)_darwin_arm cmd/$(ADMIN)/main.go

gen-proto: install-proto ## generate golang from protobuf files
	protoc -I proto/ proto/jamel/*.proto --go_out=./gen/go/ --go_opt=paths=source_relative --go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative
	
install-proto: ## install protobuf requirements
	[ -f $(BIN)/protoc-gen-go ] || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && cp $(GOPATH)/bin/protoc-gen-go $(BIN)
	[ -f $(BIN)/protoc-gen-go-grpc ] || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && cp $(GOPATH)/bin/protoc-gen-go-grpc $(BIN)
	[ -f $(BIN)/protoc ] || (curl -sSfL https://github.com/protocolbuffers/protobuf/releases/download/v29.1/protoc-29.1-linux-x86_64.zip -o $(BIN)/protoc-29.1-linux-x86_64.zip && unzip -o $(BIN)/protoc-29.1-linux-x86_64.zip && rm -rf include readme.txt)

remove-certs: ## remove old certs
	find cmd -name "*.crt" -delete
	find cmd -name "*.key" -delete

gen-certs: ## generate grpc ssl certs
	echo "subjectAltName = $(SUBJECT_ALT_NAME)" > extfile.cnf

	for name in server client admin; do \
		openssl req -newkey rsa:4096 -nodes -keyout $$name.key -out $$name.csr -subj "/CN=$$name"; \
		openssl x509 -req -sha256 -days 365 -in $$name.csr -signkey $$name.key -out $$name.crt -extfile extfile.cnf; \
	done

	cp server.key server.crt client.crt admin.crt cmd/server
	cp client.key client.crt server.crt cmd/client
	cp admin.key admin.crt server.crt cmd/admin

	rm -f *.crt *.key *.cnf *.csr

dev-up: ## up development environment
	docker compose -f docker-compose.local.yaml up -d

dev-rm: ## rm development environment
	docker compose -f docker-compose.local.yaml down
	sudo rm -rf rabbitmq
	sudo rm -rf minio

build-images: ## build Docker images with apps
	docker build -t git.codenrock.com:5050/5hm3l/jamel/$(APP)-$(CLIENT) -f Dockerfile.client .
	docker build -t git.codenrock.com:5050/5hm3l/jamel/$(APP)-$(SERVER) -f Dockerfile.server .
