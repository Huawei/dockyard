#BUILDTAGS=
#export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
		go build -tags "$(BUILDTAGS)" -o dockyard .
		make -C oss/chunkserver

install:
		cp dockyard /usr/local/bin/dockyard
clean:
		go clean
		@rm -rf oss/chunkserver/*.o
	    @rm -rf oss/chunkserver/spy_server
