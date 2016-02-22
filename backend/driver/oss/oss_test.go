package oss

import (
	"testing"

	"github.com/containerops/wrench/setting"
)

func Test_osssave(t *testing.T) {
	err := setting.SetConfig("../../../conf/containerops.conf")
	file := "oss_test.go"
	_, err = osssave(file)
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossgetfileinfo(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig("../../../conf/containerops.conf")
	err = ossgetfileinfo(path)
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossdownload(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig("../../../conf/containerops.conf")
	err = ossdownload(path, "/root/gopath/chunkserver/downloadtest")
	if err != nil {
		t.Error(err)
		return
	}
}

func Test_ossdel(t *testing.T) {
	path := "oss_test.go"
	err := setting.SetConfig("../../../conf/containerops.conf")
	err = ossdel(path)
	if err != nil {
		t.Error(err)
		return
	}
}
