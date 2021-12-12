GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GONAME = gop

default: build

workdir:
		mkdir -p workdir

build: workdir/gop-linux workdir/gop-windows

workdir/gop-linux: $(GOFILES)
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-w -s" -o workdir/$(GONAME) .

workdir/gop-windows: $(GOFILES)
		GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-w -s" -o workdir/$(GONAME).exe .

install: $(GOFILES)
		go install .

test: test-all

test-all:
		@go test -v $(GOPACKAGES)

lint: lint-all

lint-all:
		@golint -set_exit_status $(GOPACKAGES)

clean:
		rm -rf workdir/
