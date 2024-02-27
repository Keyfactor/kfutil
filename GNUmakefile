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
	VERSION := v1.2.1
endif
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
BASEDIR := ${HOME}/go/bin
INSTALLDIR := ${BASEDIR}
MARKDOWN_FILE := README.md
TEMP_TOC_FILE := temp_toc.md



default: build

build: fmt
	go install

release:
	mkdir -p ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64/kfutil
	cp README.md ./bin/${BINARY}_${VERSION}_darwin_amd64
	cp LICENSE ./bin/${BINARY}_${VERSION}_darwin_amd64
	cp CHANGELOG.md ./bin/${BINARY}_${VERSION}_darwin_amd64
	cp -r docs ./bin/${BINARY}_${VERSION}_darwin_amd64
	cd ./bin && zip ./${BINARY}_${VERSION}_darwin_amd64.zip ./${BINARY}_${VERSION}_darwin_amd64/* && cd ..
	rm -rf ./bin/${BINARY}_${VERSION}_darwin_amd64
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

install: fmt
	go build -o ${BINARY}
	rm -rf ${INSTALLDIR}/${BINARY}
	mkdir -p ${INSTALLDIR}
	chmod oug+x ${BINARY}
	cp ${BINARY} ${INSTALLDIR}
	mkdir -p ${HOME}/.local/bin || true
	mv ${BINARY} ${HOME}/.local/bin/${BINARY}

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

check_toc:
	@grep -q 'TOC_START' $(MARKDOWN_FILE) && echo "TOC already exists." || (echo "TOC not found. Generating..." && $(MAKE) generate_toc)

generate_toc:
	# Generate TOC and store in temporary file
	markdown-toc -i $(MARKDOWN_FILE) > $(TEMP_TOC_FILE)
	# check if files are different
#	@diff -q $(TEMP_TOC_FILE) $(MARKDOWN_FILE) && echo "TOC is up to date." || (echo "TOC is not up to date. Updating..." && mv $(TEMP_TOC_FILE) $(MARKDOWN_FILE))
#	@rm -f $(TEMP_TOC_FILE)


.PHONY: build prerelease release install test fmt vendor version setversion