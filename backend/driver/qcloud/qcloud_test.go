package qcloud

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func Test_qcloudsave(t *testing.T) {

	err := setting.SetConfig("../../../conf/containerops.conf")
	if err != nil {
		t.Error(err)
	}

	file := "qcloud_test.go"
	url, err := qcloudsave(file)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(url)

	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
