# What is this
This is the integration test by using dockyard client libs and client itself.
Each test file will talk to a server by Restful APIs.

## How to do integration test
### Setup the test
First, start a dockyard web server
```
$ cd $GOPATH/github.com/containerops/dockyard
$ go build
$ // modify your containerops.conf or runtime.conf
$ ./dockyard web --port=1234 
```

### API test
Then you can run `go test`
```
$ export US_TEST_SERVER=http://localhost:1234
# cd $GOPATH/github.com/containerops/dockyard
$ cd tests/integrate
$ go test
```


### Client test
Or you can run the `dc.sh` script, which is doing this:

```
$ cd $GOPATH/github.com/containerops/dockyard
$ sh ./tests/integrate/dc.sh
```

The success output will be:
```
start to run test
push file appA
--------------------------------
push file appB
--------------------------------
list appA and appB
osA/archA/appA
osB/archB/appB
--------------------------------
pull appA
start to download file:  osA/archA/appA
Congratulations! The file is valid!
success in downloading and verifing file:  /root/.dockyard/cache/n/r/osA/archA/appA
--------------------------------
delete appA
--------------------------------
list appB only
osB/archB/appB
end of the test
--------------------------------
```
