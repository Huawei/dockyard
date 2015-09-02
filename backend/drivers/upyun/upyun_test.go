package upyun

import (
	"net/http"
	"testing"

	"github.com/astaxie/beego/config"
)

func Test_upyunsave(t *testing.T) {
	var err error
	var conf config.ConfigContainer
	var url string

	conf, err = config.NewConfig("ini", "../../../conf/containerops.conf")
	if err != nil {
		t.Error(err)
	}

	d := new(UpyunDrv)
	err = d.ReadConfig(conf)
	if err != nil {
		t.Error(err)
	}

	file := "upyun_test.go"
	url, err = upyunsave(file)
	if err != nil {
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
