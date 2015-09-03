package backend

import (
	"testing"
)

func Test_amazons3save(t *testing.T) {

	file := "amazons3_test.go"
	url, err := amazons3save(file)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(url)
}
