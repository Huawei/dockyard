package qiniu

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func Test_qiniusave(t *testing.T) {
	var err error
	var url string

	if err = setting.SetConfig("../../../conf/containerops.conf"); err != nil {
		t.Error(err)
	}

	file := "qiniu_test.go"
	url, err = qiniusave(file)
	if err != nil {
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
