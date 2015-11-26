package aliyun

import (
	"net/http"
	"testing"
	
	"github.com/containerops/wrench/setting"
)

func Test_aliyunsave(t *testing.T) {
	
	var url string
	
	
	err := setting.SetConfig("../../../conf/containerops.conf")

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
