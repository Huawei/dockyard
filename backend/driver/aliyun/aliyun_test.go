package aliyun

import (
	"net/http"
	"testing"

	"github.com/containerops/wrench/setting"
)

func Test_aliyunsave(t *testing.T) {
	var err error
	var url string

	if err = setting.SetConfig("../../../conf/containerops.conf"); err != nil {
		t.Error(err)
	}

	file := "aliyun_test.go"
	url, err = aliyunsave(file)
	if err != nil {
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
