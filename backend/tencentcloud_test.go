package backend

import (
	"os"
	"testing"
)

func Test_tencentcloudsave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		t.Error("read env GOPATH fail")
		return
	}
	file := gopath + "/src/github.com/containerops/dockyard/backend/tencentcloud_test.go"
	url, err := tencentcloudsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(url)
}
