BUILDDIR ?= build
OSS ?= linux darwin freebsd windows
ARCHS ?= amd64 arm64
VERSION ?= $(shell git describe --tags `git rev-list -1 HEAD`)

build: $(BUILDDIR)/fvpn

clean:
	rm -rf "$(BUILDDIR)"

install: build

define fvpn
$(BUILDDIR)/fvpn-$(1)-$(2): export CGO_ENABLED := 0
$(BUILDDIR)/fvpn-$(1)-$(2): export GOOS := $(1)
$(BUILDDIR)/fvpn-$(1)-$(2): export GOARCH := $(2)
$(BUILDDIR)/fvpn-$(1)-$(2):
	go build \
	-ldflags="-s -w -X main.appVersion=$(VERSION) -X main.DSN=$(SENTRY_DSN) -X main.firebaseApiKey=$(PRODUCTION_FIREBASE_API_KEY) -X main.apiHost=$(PRODUCTION_API_URL)" \
	-trimpath -v -o "$(BUILDDIR)/fvpn-$(1)-$(2)"
endef
$(foreach OS,$(OSS),$(foreach ARCH,$(ARCHS),$(eval $(call fvpn,$(OS),$(ARCH)))))

$(BUILDDIR)/fvpn: $(foreach OS,$(OSS),$(foreach ARCH,$(ARCHS),$(BUILDDIR)/fvpn-$(OS)-$(ARCH)))
	@mkdir -vp "$(BUILDDIR)"

.PHONY: clean build install