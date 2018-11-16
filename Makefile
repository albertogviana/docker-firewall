.PHONY: install-dep
install-dep:
	export DEP_RELEASE_TAG=v0.4.1; curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

.PHONY: dep
dep:
	dep ensure

.PHONY: verify
verify: go_fmt go_vet go_lint test

.PHONY: build
build: 
	CGO_ENABLED=0 GOOS=linux go build -o cmd/docker-firewall/docker-firewall cmd/docker-firewall/main.go

.PHONY: test
test:
	GOCACHE=off go test -v -cover -coverprofile=coverage.out  $$(go list ./... | grep -v '/vendor/')

.PHONY: go_vet
go_vet:
	go vet -v $$(go list ./... | grep -v '/vendor/')

.PHONY: go_fmt
go_fmt:
	git ls-files '*.go' | grep -v 'vendor/' | xargs gofmt -s -w

.PHONY: go_lint
go_lint: install-golint
	golint $(go list ./... | grep -v /vendor)

.PHONY: install-golint
install-golint:
	GOLINT_CMD=$(shell command -v golint 2> /dev/null)
ifndef GOLINT_CMD
	go get golang.org/x/lint/golint
endif

.PHONY: clean-vendor
clean-vendor:
	find ./vendor -type l | xargs rm -rf