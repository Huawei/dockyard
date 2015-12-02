package backend

import (
	"testing"
)

func Test_qcloudsave(t *testing.T) {

	file := "qcloud_test.go"
	url, err := qcloudsave(file)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(url)
}
