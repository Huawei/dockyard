package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

var (
	testServer = ""
)

func init() {
	// Start a dockyard web server and set the enviornment like this:
	//     $ export US_TEST_SERVER=https://containerops.me
	testServer = os.Getenv("US_TEST_SERVER")
}

func Test_IndexHandler(t *testing.T) {
	if testServer == "" {
		fmt.Println("Skip index handler testing since 'US_TEST_SERVER' is not set")
		return
	}
	endpoint := testServer

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Errorf("Test REST API \"/\" Error: %s .", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Test REST API \"/\" Error StatusCode: %d .", resp.StatusCode)
	} else {
		t.Log("Test REST API \"/\" Successfully.")
	}
}
