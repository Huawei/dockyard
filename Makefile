BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
		go build -tags "$(BUILDTAGS)" -o dockyard .

install:
		cp dockyard /usr/local/bin/dockyard
clean:
		go clean
