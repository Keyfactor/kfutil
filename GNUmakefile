PROVIDER_DIR := $(PWD)
TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=keyfactor.com
GOFMT_FILES  := $$(find $(PROVIDER_DIR) -name '*.go' |grep -v vendor)
NAMESPACE=keyfactor
WEBSITE_REPO=https://github.com/Keyfactor/kfutil
NAME=kfutil
BINARY=${NAME}
VERSION := $(GITHUB_REF_NAME)
ifeq ($(VERSION),)
	VERSION := $(shell git tag -l | tail -n 1)
endif
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
BASEDIR := ${HOME}/go/bin
INSTALLDIR := ${BASEDIR}

default: build

build: fmt
	go install

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: fmt setversion
	go build -o ${BINARY}
	rm -rf ${INSTALLDIR}/${BINARY}
	mkdir -p ${INSTALLDIR}
	chmod oug+x ${BINARY}
	cp ${BINARY} ${INSTALLDIR}
	mv ${BINARY} /usr/local/bin/${BINARY}

vendor:
	go mod vendor

version:
	@echo ${VERSION}

setversion:
	sed -i '' -e 's/VERSION = ".*"/VERSION = "$(VERSION)"/' pkg/version/version.go

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

fmt:
	gofmt -w $(GOFMT_FILES)

prerelease: fmt setversion
	git tag -d $(VERSION) || true
	git push origin :$(VERSION) || true
	git tag $(VERSION)
	git push origin $(VERSION)

.PHONY: build prerelease release install test fmt vendor version setversion