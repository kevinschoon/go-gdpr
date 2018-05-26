BASEPATH="github.com/greencase/go-gdpr"

.PHONY: \
	all \
	dep \
	test 


all: test

dep:
	cd $$GOPATH/src/${BASEPATH} && $@ ensure

test:
	@go $@ -v -bench .
	@go vet .
