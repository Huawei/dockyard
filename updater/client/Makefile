PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin

all:
	go build -tags "$(BUILDTAGS)" -o duc .

.PHONY: man
man:
	go-md2man -in "man/duc.1.md" -out "duc.1"

install: man
	install -d -m 755 $(BINDIR)
	install -m 755 duc $(BINDIR)
	install -d -m 755 $(PREFIX)/share/man/man1
	install -m 644 *.1 $(PREFIX)/share/man/man1
	install -d -m 755 $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/bash/duc $(PREFIX)/share/bash-completion/completions

clean:
	rm -f duc duc.1

.PHONY: test .gofmt .govet .golint

test: .gofmt .govet .golint

.gofmt:
	go fmt ./...

.govet:
	go vet -x ./...

.golint:
	golint ./...

