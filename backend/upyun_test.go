package backend

import (
	"net/http"
	"os"
	"testing"
)

func Test_upcloudsave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		t.Error("read env GOPATH fail")
		return
	}

	file := gopath + "/src/github.com/containerops/dockyard/backend/upyun.go"
	url, err := upcloudsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
