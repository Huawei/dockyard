package backend

import (
	"encoding/json"
	"os"
	"testing"
)

func Test_backend_put(t *testing.T) {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		t.Error("read env GOPATH fail")
		return
	}
	file := gopath + "/src/github.com/containerops/dockyard/backend/backend_test.go"

	in := &In{Key: "asdf8976485r32r613879rwegfuiwet739ruwef", Uploadfile: file}
	jsonIn, err := json.Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	sc := NewShareChannel()
	sc.Open()

	for i := 0; i < 2; i++ {
		sc.PutIn(string(jsonIn))
	}
	sc.Close()

	for f := true; f; {
		select {
		case obj := <-sc.OutSuccess:
			t.Log(obj)
		case obj := <-sc.OutFailure:
			t.Error(obj)
		default:
			f = false
		}
	}

}
