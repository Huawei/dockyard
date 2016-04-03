package qcloud

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var testconf = "../../conf/containerops.conf"

func Test_qcloudsave(t *testing.T) {

	err := setting.SetConfig(testconf)
	if err != nil {
		t.Error(err)
	}

	file := "qcloud_test.go"
	q := new(qclouddesc)
	url, err := q.Save(file)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(url)

	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
