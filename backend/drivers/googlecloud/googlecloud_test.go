package googlecloud

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/containerops/wrench/setting"
)

var (
	upFileName   string = "/tmp/gcs_test.txt"
	downFileName string = "/tmp/new_gcs_test.txt"
	fileContent  string = "Just for test gcs.\n Congratulations! U are sucess."
)

func newTestFile(t *testing.T) (f *os.File, err error) {
	file, err := os.Create(upFileName)
	if err != nil {
		t.Error(err)
	}

	ret, err := file.WriteString(fileContent)
	if err != nil {
		t.Error(err)
		t.Fatalf("GCS_TEST Write String ret =  %v", ret)
	}
	if err != nil {
		return nil, err
	} else {
		return file, nil
	}
}

// Unit Test for gcs
func TestGcssave(t *testing.T) {
	file, err := newTestFile(t)
	if err != nil {
		t.Error(err)
	}

	err = setting.SetConfig("../../../conf/containerops.conf")
	if err != nil {
		t.Error(err)
	}
	retUrl, err := googlecloudsave(upFileName)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get(retUrl)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	// Open file for writing
	nFile, err := os.Create(downFileName)
	if err != nil {
		t.Error(err)
	}

	// Use io.Copy to copy a file from URL to a locald disk
	_, err = io.Copy(nFile, resp.Body)
	if err != nil {
		t.Error(err)
	}

	buf, err := ioutil.ReadFile(downFileName)
	if err != nil {
		t.Error(err)
	}
	file.Close()

	isEqual := strings.EqualFold(fileContent, string(buf))
	if !isEqual {
		t.Fatalf("Testing fail, content of uploadFile is not the same as the content of downloadFile")
	}
}
