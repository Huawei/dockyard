package oss

import (
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var testconf = "../../conf/containerops.conf"

func Test_osssave(t *testing.T) {
	err := setting.SetConfig(testconf)
	file := "oss_test.go"
	o := new(ossdesc)
	_, err = o.Save(file)
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossgetfileinfo(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig(testconf)
	o := new(ossdesc)
	err = o.Get(path)
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossdownload(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig(testconf)
	o := new(ossdesc)
	err = o.Download(path, "/root/gopath/chunkserver/downloadtest")
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossdel(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig(testconf)
	o := new(ossdesc)
	err = o.Delete(path)
	if err != nil {
		t.Error(err)
		return
	}
}
