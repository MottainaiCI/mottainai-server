NAME ?= mottainai-server
PACKAGE_NAME ?= $(NAME)
PACKAGE_CONFLICT ?= $(PACKAGE_NAME)-beta
REVISION := $(shell git rev-parse --short HEAD || echo unknown)
VERSION := $(shell git describe --tags || cat pkg/settings/settings.go | echo $(REVISION) || echo dev)
VERSION := $(shell echo $(VERSION) | sed -e 's/^v//g')
ITTERATION := $(shell date +%s)
BUILD_PLATFORMS ?= -osarch="linux/amd64" -osarch="linux/386"
SUBDIRS =
DESTDIR =
UBINDIR ?= /usr/bin
LIBDIR ?= /usr/lib
SBINDIR ?= /sbin
USBINDIR ?= /usr/sbin
BINDIR ?= /bin
LIBEXECDIR ?= /usr/libexec
SYSCONFDIR ?= /etc
LOCKDIR ?= /var/lock
LIBDIR ?= /var/lib
EXTENSIONS ?=

all: deps multiarch-build install

build-test: test multiarch-build

help:
	# make all => deps test lint build
	# make deps - install all dependencies
	# make test - run project tests
	# make lint - check project code style
	# make build - build project for all supported OSes

clean:
	rm -rf vendor/
	rm -rf release/

deps:
	go env
	# Installing dependencies...
	go get golang.org/x/lint/golint
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls

build:
ifeq ($(EXTENSIONS),)
		go build
else
		go build -tags $(EXTENSIONS)
endif

multiarch-build:
ifeq ($(EXTENSIONS),)
		# Building gitlab-ci-multi-runner for $(BUILD_PLATFORMS)
		gox $(BUILD_PLATFORMS) -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition"
		CC="arm-linux-gnueabi-gcc" LD_LIBRARY_PATH=/usr/arm-linux-gnueabi/lib gox -osarch="linux/arm" -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition"
else
		# Building gitlab-ci-multi-runner for $(BUILD_PLATFORMS)
		gox $(BUILD_PLATFORMS) -tags $(EXTENSIONS) -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition" -parallel 1 -cgo
		CC="arm-linux-gnueabi-gcc" LD_LIBRARY_PATH=/usr/arm-linux-gnueabi/lib gox -tags $(EXTENSIONS) -osarch="linux/arm" -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition" -parallel 1 -cgo
endif

lint:
	# Checking project code style...
	golint ./... | grep -v "be unexported"

test:
	# Running tests... ${TOTEST}
	go test -v -tags all -cover -race ./...

install:
	install -d $(DESTDIR)$(LOCKDIR)
	install -d $(DESTDIR)$(BINDIR)
	install -d $(DESTDIR)$(UBINDIR)
	install -d $(DESTDIR)$(SYSCONFDIR)
	install -d $(DESTDIR)$(LIBDIR)

	install -d $(DESTDIR)$(LOCKDIR)/mottainai
	install -d $(DESTDIR)$(SYSCONFDIR)/mottainai
	install -d $(DESTDIR)$(LIBDIR)/mottainai

	install -m 0755 $(NAME) $(DESTDIR)$(UBINDIR)/
	cp -rf templates/ $(DESTDIR)$(LIBDIR)/mottainai
	cp -rf public/ $(DESTDIR)$(LIBDIR)/mottainai

	install -m 0755 contrib/config/mottainai-server.yaml.example $(DESTDIR)$(SYSCONFDIR)/mottainai/
