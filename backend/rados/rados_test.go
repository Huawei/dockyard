package rados

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var (
	r = &radosdesc{}
	RadosBinary	 string = "rados"
	username 	 string = "client."
	poolname	 string = ""
	upFileName   string = "/tmp/rados_test.txt"
	fileContent  string = "Just for test rados.\n Congratulations! U are sucess."
)

func TestFileInit(t *testing.T) {
	file, err := os.Create(upFileName)
	if err != nil {
		t.Error(err)
	}

	ret, err := file.WriteString(fileContent)
	if err != nil {
		t.Error(err)
		t.Fatalf("RADOS_TEST Write String ret =  %v", ret)
	}

	if err = setting.SetConfig("../../conf/containerops.conf"); err != nil {
		t.Error(err)
	}
	username += setting.Username
	poolname += setting.Poolname
}

//Unit test for rados
func TestRadosSave(t *testing.T) {
	var err error

	_, err = r.Save(upFileName)
	if err != nil {
		t.Error(err)
	}

	//Print all object in pool
	buf, err := exec.Command(RadosBinary, "-p", poolname, "ls", "-n", username).CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	t.Log(string(buf))
}

func TestRadosGet(t *testing.T) {
	var err error

	buf, err := r.Get(upFileName)
	if err != nil {
		t.Error(err)
	}

	isEqual := strings.EqualFold(fileContent, string(buf))
	if !isEqual {
		t.Fatalf("Testing fail, content of uploadFile is not the same as the content of downloadFile")
	}
}

func TestRadosDelete(t *testing.T) {
	var err error

	err = r.Delete(upFileName)
	if err != nil {
		t.Error(err)
	}

	//Print all object in pool
	buf, err := exec.Command(RadosBinary, "-p", poolname, "ls", "-n", username).CombinedOutput() 
	if err != nil {
		t.Error(err)
	}
	t.Log(string(buf))
}