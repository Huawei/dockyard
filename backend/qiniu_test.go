package backend

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/astaxie/beego/config"
)

func getqiniuconfile(file string) (err error) {
	var tmperr error
	var conf config.ConfigContainer

	conf, tmperr = config.NewConfig("ini", file)
	if tmperr != nil {
		return tmperr
	}

	DRIVER = "qiniu"
	if DRIVER == "" {
		return errors.New("read config file's backenddriver failed!")
	}

	ENDPOINT = conf.String(DRIVER + "::endpoint")
	if ENDPOINT == "" {
		return errors.New("read config file's endpoint failed!")
	}

	BUCKETNAME = conf.String(DRIVER + "::bucket")
	if BUCKETNAME == "" {
		return errors.New("read config file's bucket failed!")
	}

	AccessKeyID = conf.String(DRIVER + "::accessKeyID")
	if AccessKeyID == "" {
		return errors.New("read config file's accessKeyID failed!")
	}

	AccessKeySecret = conf.String(DRIVER + "::accessKeysecret")
	if AccessKeySecret == "" {
		return errors.New("read config file's accessKeysecret failed!")
	}

	return nil
}

func Test_qiniusave(t *testing.T) {

	var gopath string
	gopath = os.Getenv("GOPATH")
	conffile := gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf"
	var err error
	err = getqiniuconfile(conffile)
	if nil != err {
		fmt.Printf("read conf/runtime.conf error: %v", err)
	}

	file := gopath + "/src/github.com/containerops/dockyard/backend/qiniu.go"
	var url string
	url, err = qiniusave(file)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}

}
