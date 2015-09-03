package backend

import (
	"net/http"
	"testing"
)

func Test_aliyunsave(t *testing.T) {

	file := "aliyun_test.go"
	url, err := aliyunsave(file)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = http.Get(url)
	if err != nil {
		t.Error(err)
	}
}
