version=''
gitCommit=$(shell git rev-parse --verify HEAD)

export GO111MODULE=on

.PHONY: dep
dep:
	go mod download

.PHONY: mod_vendor
mod_vendor:
	go mod vendor

.PHONY: verify
verify: go_fmt go_vet go_lint test

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -v -ldflags '-X "main.version=${version}" -X "main.gitCommit=${gitCommit}"' -o cmd/docker-firewall/docker-firewall cmd/docker-firewall/main.go

.PHONY: test
test:
	# Here sudo -E env "PATH=$PATH" make test is required for running tests with
  	# sudo permissions since it is testing iptables, sudo or root permissions are required.
	sudo -E env PATH="$(PATH)" go test -v -cover -coverprofile=coverage.out  $$(go list ./... | grep -v '/vendor/')

.PHONY: go_vet
go_vet:
	go vet -v $$(go list ./... | grep -v '/vendor/')

.PHONY: go_fmt
go_fmt:
	git ls-files '*.go' | grep -v 'vendor/' | xargs gofmt -s -w

.PHONY: go_lint
go_lint: install-golint
	golint $$(go list ./... | grep -v /vendor)

.PHONY: install-golint
install-golint:
	GOLINT_CMD=$(shell command -v golint 2> /dev/null)
ifndef GOLINT_CMD
	go get golang.org/x/lint/golint
endif

.PHONY: clean-vendor
clean-vendor:
	find ./vendor -type d | xargs rm -rf