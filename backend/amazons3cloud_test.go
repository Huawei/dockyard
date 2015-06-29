package backend

import (
	"os"
	"testing"
)

func Test_amazons3cloudsave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		t.Error("read env GOPATH fail")
		return
	}
	file := gopath + "/src/github.com/containerops/dockyard/backend/amazons3cloud_test.go"
	url, err := amazons3cloudsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(url)
}
