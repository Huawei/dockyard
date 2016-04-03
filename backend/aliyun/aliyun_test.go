package aliyun

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var testconf = "../../conf/containerops.conf"

func Test_aliyunsave(t *testing.T) {
	var err error
	var url string

	if err = setting.SetConfig(testconf); err != nil {
		t.Error(err)
	}

	file := "aliyun_test.go"
	a := new(aliyundesc)
	url, err = a.Save(file)
	if err != nil {
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
