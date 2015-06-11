package backend

import (
	"fmt"
	"os"
	"testing"
)

func init() {

	var gopath string
	gopath = os.Getenv("GOPATH")
	conffile := gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf"
	var err error
	err = getconfile(conffile)
	if nil != err {
		fmt.Printf("read conf/runtime.conf error: %v", err)
	}

}

func Test_backend_put(t *testing.T) {
	const jsonInput = `{ 
	"key" : "asdf8976485r32r613879rwegfuiwet739ruwef" ,
	"uploadfile" : "/home/lgp/1.txt" }`

	sc := NewShareChannel(5)
	sc.StartRoutine()
	for i := 0; i < 10; i++ {
		sc.PutIn(jsonInput)
	}

}
