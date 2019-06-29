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
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

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
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/maxbrunsfeld/counterfeiter
	go get -u github.com/onsi/gomega/...

build:
ifeq ($(EXTENSIONS),)
		go build
else
		go build -tags $(EXTENSIONS)
endif

multiarch-build:
ifeq ($(EXTENSIONS),)
		gox $(BUILD_PLATFORMS) -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition"
		CC="arm-linux-gnueabi-gcc" LD_LIBRARY_PATH=/usr/arm-linux-gnueabi/lib gox -osarch="linux/arm" -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition"
else
		gox $(BUILD_PLATFORMS) -tags $(EXTENSIONS) -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition" -parallel 1 -cgo
		CC="arm-linux-gnueabi-gcc" LD_LIBRARY_PATH=/usr/arm-linux-gnueabi/lib gox -tags $(EXTENSIONS) -osarch="linux/arm" -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}" -ldflags "-extldflags=-Wl,--allow-multiple-definition" -parallel 1 -cgo
endif

lint:
	golint ./... | grep -v "be unexported"

test:
	go test -v -tags all -cover -race ./...

ginkgo-test:
	ginkgo -p -r --randomizeAllSpecs -failOnPending --trace

docker-test:
	docker run -v $(ROOT_DIR)/:/test \
	-e ACCEPT_LICENSE=* \
	--entrypoint /bin/bash -ti --user root --rm mottainaici/test -c \
	"mkdir -p /root/go/src/github.com/MottainaiCI && \
	cp -rf /test /root/go/src/github.com/MottainaiCI/mottainai-server && \
	cd /root/go/src/github.com/MottainaiCI/mottainai-server && \
	make deps test"

compose-test-run: build
		@tmpdir=`mktemp --tmpdir -d`; \
		cp -rf $(ROOT_DIR)/contrib/docker-compose "$$tmpdir"; \
		pushd "$$tmpdir/docker-compose"; \
		mv docker-compose.arangodb.yml docker-compose.yml; \
		trap 'docker-compose down -v --remove-orphans;rm -rf "$$tmpdir"' EXIT; \
		echo ">> Server will be avilable at: http://127.0.0.1:4545" ; \
		sed -i "s|#- ./mottainai-server.yaml:/etc/mottainai/mottainai-server.yaml|- "$(ROOT_DIR)"/mottainai-server:/usr/bin/mottainai-server|g" docker-compose.yml; \
		sed -i "s|# For static config:|- "$(ROOT_DIR)":/var/lib/mottainai|g" docker-compose.yml; \
		docker-compose up

kubernetes:
	make/kubernetes

helm-gen:
	make/helm-gen

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

gen-fakes:
	counterfeiter -o tests/fakes/http_client.go pkg/client/client.go HttpClient