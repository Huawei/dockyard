package backend

import (
	"net/http"
	"testing"

	"github.com/astaxie/beego/config"
)

func Test_aliyunsave(t *testing.T) {
	var err error
	var conf config.ConfigContainer
	var url string
	//var d *AliyunDrv

	conf, err = config.NewConfig("ini", "../../../conf/containerops.conf")
	if err != nil {
		t.Error(err)
	}

	d := new(AliyunDrv)
	err = d.ReadConfig(conf)
	if err != nil {
		t.Error(err)
	}

	file := "aliyun_test.go"
	url, err = aliyunsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
