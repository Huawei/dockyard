package backend

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func Test_qiniusave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")

	conffile := gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf"
	var err error
	err = getconfile(conffile)
	if nil != err {
		fmt.Printf("读取配置文件 conf/runtime.conf 错误: %v", err)
	}

	DRIVER = "qiniu"

	file := gopath + "/src/github.com/containerops/dockyard/backend/1.txt"
	var url string
	url, err = qiniusave(file)
	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}

}
