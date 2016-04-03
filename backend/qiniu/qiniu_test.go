package qiniu

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var testconf = "../../conf/containerops.conf"

func Test_qiniusave(t *testing.T) {
	var err error
	var url string

	if err = setting.SetConfig(testconf); err != nil {
		t.Error(err)
	}

	file := "qiniu_test.go"
	q := new(qiniudesc)
	url, err = q.Save(file)
	if err != nil {
		t.Error(err)
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
