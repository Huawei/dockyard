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
)

func loadTestData(t *testing.T) dus_utils.DyUpdaterServerStorage {
	var local DyUpdaterServerLocal
	_, path, _, _ := runtime.Caller(0)
	realPath := filepath.Join(filepath.Dir(path), "testdata")

	l, err := local.New(localPrefix + ":/" + realPath)
	assert.Nil(t, err, "Fail to setup a local test storage")

	return l
}

// TestBasic
func TestBasic(t *testing.T) {
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

func TestList(t *testing.T) {
	l := loadTestData(t)
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

func TestPut(t *testing.T) {
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

	meta, err := l.GetMeta(validKey)
	assert.Nil(t, err, "Fail to get meta data")

	getData, err := l.Get(validKey, meta.GetHash())
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(getData), testData, "Fail to get correct file")
}

func TestGet(t *testing.T) {
	l := loadTestData(t)

	key := "containerops/official/appA"
	invalidKey := "containerops/official"

	_, err := l.GetMeta(invalidKey)
	assert.Equal(t, err, dus_utils.ErrorsDUSSInvalidKey)
	meta, err := l.GetMeta(key)
	assert.Nil(t, err, "Fail to load meta data")

	_, err = l.Get(invalidKey, "invalid hash")
	assert.Equal(t, err, dus_utils.ErrorsDUSSInvalidKey)

	data, err := l.Get(key, meta.Hash)
	assert.Nil(t, err, "Fail to load file")
	assert.Equal(t, string(data), "this is test appA", "Fail to get correct file")
}
