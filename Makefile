BINARY:=raknet-proxy
GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o build/$(BINARY)-$(GOOS)-$(GOARCH) ./cmd/$(BINARY)

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -rf build $(BINARY) *.pcap
