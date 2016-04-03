package upyun

import (
	"net/http"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var testconf = "../../conf/containerops.conf"

func Test_upyunsave(t *testing.T) {
	var err error
	var url string

	if err = setting.SetConfig(testconf); err != nil {
		t.Error(err)
	}

	file := "upyun_test.go"
	u := new(upyundesc)
	url, err = u.Save(file)
	if err != nil {
		t.Error(err)
	}

	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
