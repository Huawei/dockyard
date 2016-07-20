PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin

all:
	go build -tags "$(BUILDTAGS)" -o dus .

.PHONY: man
man:
	go-md2man -in "man/dus.1.md" -out "dus.1"

install: man
	install -d -m 755 $(BINDIR)
	install -m 755 dus $(BINDIR)
	install -d -m 755 $(PREFIX)/share/man/man1
	install -m 644 *.1 $(PREFIX)/share/man/man1
	install -d -m 755 $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/bash/dus $(PREFIX)/share/bash-completion/completions

uninstall:
	rm -f $(PREFIX)/share/man/man1/dus.1 $(PREFIX)/share/bash-completion/completions/dus
clean:
	rm -f dus dus.1

.PHONY: test .gofmt .govet .golint

test: .gofmt .govet .golint

.gofmt:
	go fmt ./...

.govet:
	go vet -x ./...

.golint:
	golint ./...

