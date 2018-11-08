install-dep:
	export DEP_RELEASE_TAG=v0.4.1; curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

dep:
	dep ensure

test:
	go test -v -cover $$(go list ./... | grep -v '/vendor/')

vet:
	go vet -v $$(go list ./... | grep -v '/vendor/')

clean-vendor:
	find ./vendor -type l | xargs rm -rf