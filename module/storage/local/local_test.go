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
package local

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/containerops/dockyard/module"
	_ "github.com/containerops/dockyard/module/km/local"
)

func loadTestData(t *testing.T) (module.UpdateServiceStorage, string) {
	var local UpdateServiceStorageLocal
	_, path, _, _ := runtime.Caller(0)
	realPath := filepath.Join(filepath.Dir(path), "testdata")

	km := "local:/" + realPath
	l, err := local.New(localPrefix+":/"+realPath, km)
	assert.Nil(t, err, "Fail to setup a local test storage")

	return l, realPath
}

// TestBasic
func TestLocalBasic(t *testing.T) {
	var local UpdateServiceStorageLocal

	validURL := "local://tmp/containerops_storage_cache"
	ok := local.Supported(validURL)
	assert.Equal(t, ok, true, "Fail to get supported status")
	ok = local.Supported("localInvalid://tmp/containerops_storage_cache")
	assert.Equal(t, ok, false, "Fail to get supported status")

	l, err := local.New(validURL, "")
	assert.Nil(t, err, "Fail to setup a local storage")
	assert.Equal(t, l.String(), validURL)
}

func TestLocalList(t *testing.T) {
	l, _ := loadTestData(t)
	key := "containerops/official"
	validCount := 0

	apps, _ := l.List(key)
	for _, app := range apps {
		if app == "appA" || app == "appB" {
			validCount++
		}
	}
	assert.Equal(t, validCount, 2, "Fail to get right apps")
}

func TestLocalPut(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "us-test-")
	defer os.RemoveAll(tmpPath)
	assert.Nil(t, err, "Fail to create temp dir")

	testData := "this is test DATA, you can put in anything here"

	var local UpdateServiceStorageLocal
	l, err := local.New(localPrefix+":/"+tmpPath, localPrefix+":/"+tmpPath)
	assert.Nil(t, err, "Fail to setup local repo")

	invalidKey := "containerops/official"
	err = l.Put(invalidKey, []byte(testData))
	assert.NotNil(t, err, "Fail to put with invalid key")

	validKey := "containerops/official/appA"
	err = l.Put(validKey, []byte(testData))
	assert.Nil(t, err, "Fail to put key")

	_, err = l.GetMeta("containerops/official")
	assert.Nil(t, err, "Fail to get meta data")

	getData, err := l.Get(validKey)
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(getData), testData, "Fail to get correct file")
}

func TestLocalGet(t *testing.T) {
	l, kmPath := loadTestData(t)

	key := "containerops/official"
	invalidKey := "containerops/official/invalid"

	defer os.RemoveAll(filepath.Join(kmPath, key, defaultKeyDir))
	_, err := l.GetPublicKey(key)
	assert.Nil(t, err, "Fail to load public key")
	_, err = l.GetMetaSign(key)
	assert.Nil(t, err, "Fail to load  sign file")

	_, err = l.GetMeta(invalidKey)
	assert.NotNil(t, err, "Fail to get meta from invalid key")
	_, err = l.GetMeta(key)
	assert.Nil(t, err, "Fail to load meta data")

	_, err = l.Get("invalidinput")
	assert.NotNil(t, err, "Fail to get by invalid key")

	data, err := l.Get(key + "/appA")
	expectedData := "This is the content of appA."
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(data), expectedData, "Fail to get correct file")
}
