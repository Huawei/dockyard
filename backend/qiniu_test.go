package backend

import (
	"net/http"
	"os"
	"testing"

	"github.com/qiniu/api/conf"
)

/*
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
*/
func Test_qiniucloudsave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		t.Error("read env GOPATH fail")
		return
	}

	conf.ACCESS_KEY = g_qiniuAccessKeyID
	conf.SECRET_KEY = g_qiniuAccessKeySecret

	file := gopath + "/src/github.com/containerops/dockyard/backend/qiniu.go"
	url, err := qiniucloudsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
