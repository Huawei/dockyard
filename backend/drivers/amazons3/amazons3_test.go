package amazons3

import (
	"testing"
	
	"github.com/containerops/wrench/setting"
)

func Test_amazons3save(t *testing.T) {
	var url string
	
	
	err := setting.SetConfig("../../../conf/containerops.conf")
	
	file := "amazons3_test.go"
	url, err = amazons3save(file)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(url)
}
