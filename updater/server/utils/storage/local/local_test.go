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

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
	_ "github.com/containerops/dockyard/updater/server/utils/km/local"
)

func loadTestData(t *testing.T) (dus_utils.DyUpdaterServerStorage, string) {
	var local DyUpdaterServerLocal
	_, path, _, _ := runtime.Caller(0)
	realPath := filepath.Join(filepath.Dir(path), "testdata")

	l, err := local.New(localPrefix + ":/" + realPath)
	assert.Nil(t, err, "Fail to setup a local test storage")

	return l, realPath
}

// TestBasic
func TestLocalBasic(t *testing.T) {
	var local DyUpdaterServerLocal

	validURL := "local://tmp/containerops_storage_cache"
	ok := local.Supported(validURL)
	assert.Equal(t, ok, true, "Fail to get supported status")
	ok = local.Supported("localInvalid://tmp/containerops_storage_cache")
	assert.Equal(t, ok, false, "Fail to get supported status")

	l, err := local.New(validURL)
	assert.Nil(t, err, "Fail to setup a local storage")
	assert.Equal(t, l.String(), validURL)
}

func TestLocalList(t *testing.T) {
	l, _ := loadTestData(t)
	key := "containerops/official"
	validCount := 0

	apps, _ := l.List(key)
	for _, app := range apps {
		if app == "appA" {
			validCount++
		} else if app == "appB" {
			validCount++
		}
	}
	assert.Equal(t, validCount, 2, "Fail to get right apps")
}

func TestLocalPut(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "dus-test-")
	defer os.RemoveAll(tmpPath)
	assert.Nil(t, err, "Fail to create temp dir")

	testData := "this is test DATA, you can put in anything here"

	var local DyUpdaterServerLocal
	l, err := local.New(localPrefix + ":/" + tmpPath)
	assert.Nil(t, err, "Fail to setup local repo")

	invalidKey := "containerops/official"
	err = l.Put(invalidKey, []byte(testData))
	assert.Equal(t, err, dus_utils.ErrorsDUSSInvalidKey)

	validKey := "containerops/official/appA"
	err = l.Put(validKey, []byte(testData))
	assert.Nil(t, err, "Fail to put key")

	metas, err := l.GetMeta("containerops/official")
	assert.Nil(t, err, "Fail to get meta data")
	assert.Equal(t, len(metas), 1, "Fail to get meta data count")

	getData, err := l.Get(validKey)
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(getData), testData, "Fail to get correct file")
}

func TestLocalGet(t *testing.T) {
	l, kmPath := loadTestData(t)

	key := "containerops/official"
	invalidKey := "containerops/official/invalid"

	l.SetKM("local:/" + kmPath)
	_, err := l.GetPublicKey(key)
	assert.Nil(t, err, "Fail to load public key")
	_, err = l.GetMetaSign(key)
	assert.Nil(t, err, "Fail to load  sign file")

	_, err = l.GetMeta(invalidKey)
	assert.NotNil(t, err, "Fail to get meta from invalid key")
	metas, err := l.GetMeta(key)
	assert.Nil(t, err, "Fail to load meta data")
	assert.Equal(t, len(metas), 2, "Fail to get meta data count")

	_, err = l.Get("invalidinput")
	assert.Equal(t, err, dus_utils.ErrorsDUSSInvalidKey)

	data, err := l.Get(key + "/appA")
	expectedData := "This is the content of appA."
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(data), expectedData, "Fail to get correct file")
}
