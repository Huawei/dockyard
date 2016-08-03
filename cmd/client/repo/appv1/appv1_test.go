/*
Copyright 2016 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package appV1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/containerops/dockyard/cmd/client/module"
	"github.com/containerops/dockyard/utils"
)

func getTestURL() string {
	// Start a dus server and set the enviornment like this:
	//     $ export US_TEST_SERVER=appV1://localhost:1234
	server := os.Getenv("US_TEST_SERVER")
	if server == "" {
		return ""
	}

	return server + "/namespaceonlyfortest/repoonlyfortest"
}

// TestInitConfig tests the Init function
func TestInitConfig(t *testing.T) {
	var appV1 UpdateClientAppV1Repo

	invalidURL := "appInvalid://containerops.me/containerops/official"
	_, err := appV1.New(invalidURL)
	assert.Equal(t, err, module.ErrorsUCRepoInvalid, "Fail to parse invalid url")

	invalidURL2 := "appV1://containerops.me/containerops"
	_, err = appV1.New(invalidURL2)
	assert.Equal(t, err, module.ErrorsUCRepoInvalid, "Fail to parse invalid url")

	validURL := "appV1://containerops.me/containerops/official"
	f, err := appV1.New(validURL)
	assert.Nil(t, err, "Fail to setup a valid repo")
	assert.Equal(t, appV1.generateURL(), "http://"+"containerops.me/app/v1/containerops/official", "Fail to compose a url")
	assert.Equal(t, f.String(), validURL, "Fail to parse url")

}

// TestOper tests add/get/getmeta/getmetasign/list
func TestOper(t *testing.T) {
	var appV1 UpdateClientAppV1Repo

	validURL := getTestURL()

	// Skip the test if the testing enviornment is not ready
	if validURL == "" {
		fmt.Printf("Skip the '%s' test since the testing enviornment is not ready.\n", "List")
		return
	}

	f, _ := appV1.New(validURL)

	// Init the data and also test the put function
	_, path, _, _ := runtime.Caller(0)
	for _, n := range []string{"appA", "appB"} {
		file := filepath.Join(filepath.Dir(path), "testdata", n)
		content, _ := ioutil.ReadFile(file)
		err := f.Put(n, content)
		assert.Nil(t, err, "Fail to put file")
	}

	// Test list
	l, err := f.List()
	assert.Nil(t, err, "Fail to list")
	assert.Equal(t, len(l), 2, "Fail to list or something wrong in put")
	ok := (l[0] == "appA" && l[1] == "appB") || (l[0] == "appB" && l[1] == "appA")
	assert.Equal(t, true, ok, "Fail to list the correct data")

	// Test get file
	fileBytes, err := f.GetFile("appA")
	assert.Nil(t, err, "Fail to get file")
	expectedBytes, _ := ioutil.ReadFile(filepath.Join(filepath.Dir(path), "testdata", "appA"))
	assert.Equal(t, fileBytes, expectedBytes, "Fail to get the correct data")

	// Test get meta
	metaBytes, err := f.GetMeta()
	assert.Nil(t, err, "Fail to get meta file")

	// Test get metasign
	signBytes, err := f.GetMetaSign()
	assert.Nil(t, err, "Fail to get meta signature file")

	// Test get public key
	pubkeyBytes, err := f.GetPublicKey()
	assert.Nil(t, err, "Fail to get public key file")

	// VIP: Verify meta/sign with public to make real sure that everything works perfect
	err = utils.SHA256Verify(pubkeyBytes, metaBytes, signBytes)
	assert.Nil(t, err, "Fail to verify the meta data")
}
