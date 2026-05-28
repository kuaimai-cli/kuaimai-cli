.PHONY: build fetch_meta vendor clean dist

# 强制使用国内代理（覆盖 shell 里可能指向 proxy.golang.org 的 GOPROXY）
GO_ENV := GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE := $(shell date +%Y-%m-%d)
LDFLAGS := -s -w -X github.com/kuaimai/kuaimai-cli/internal/build.Version=$(VERSION) -X github.com/kuaimai/kuaimai-cli/internal/build.Date=$(DATE)

fetch_meta:
	@./scripts/fetch_meta/fetch_meta.sh

# 首次或依赖变更时：make vendor（需能访问 goproxy.cn）
vendor:
	$(GO_ENV) go mod vendor

build: fetch_meta
	@if [ -d vendor ]; then \
		$(GO_ENV) go build -mod=vendor -ldflags "$(LDFLAGS)" -o kuaimai-cli .; \
	else \
		$(GO_ENV) go build -ldflags "$(LDFLAGS)" -o kuaimai-cli .; \
	fi

clean:
	rm -f kuaimai-cli

dist:
	$(GO_ENV) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o dist/kuaimai-cli-linux-amd64 .
	$(GO_ENV) CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o dist/kuaimai-cli-linux-arm64 .
	$(GO_ENV) CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o dist/kuaimai-cli-windows-amd64.exe .
	$(GO_ENV) CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o dist/kuaimai-cli-darwin-amd64 .
	$(GO_ENV) CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o dist/kuaimai-cli-darwin-arm64 .
