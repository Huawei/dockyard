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

	_ "github.com/containerops/dockyard/updater/server/utils/km/local"
)

// TestRepoBasic
func TestRepoBasic(t *testing.T) {
	topDir, err := ioutil.TempDir("", "dus-repo-test-")
	defer os.RemoveAll(topDir)
	assert.Nil(t, err, "Fail to create temp dir")

	invalidURLs := []string{"a", "a/b/c"}
	for _, invalidURL := range invalidURLs {
		_, err := NewRepo(topDir, invalidURL)
		assert.NotNil(t, err, "Fail to return error while setup an invalid url")
	}

	// new
	validURL := "containerops/official"
	r, err := NewRepo(topDir, validURL)
	assert.Nil(t, err, "Fail to setup a valid url")
	assert.Equal(t, r.GetTopDir(), filepath.Join(topDir, validURL), "Fail to get the correct top dir")
	assert.Equal(t, r.GetMetaFile(), filepath.Join(topDir, validURL, defaultMeta), "Fail to get the default meta file")

	_, kmpath, _, _ := runtime.Caller(0)
	realKMPath := filepath.Join(filepath.Dir(kmpath), "testdata")
	err = r.SetKM("local:/" + realKMPath)
	assert.Nil(t, err, "Fail to set key manager")

	// add
	testData := map[string]string{
		"appA": "This is the content of appA.",
		"appB": "This is the content of appB.",
	}

	for name, value := range testData {
		err := r.Add(name, []byte(value))
		assert.Nil(t, err, "Fail to add a file")
	}

	// list
	names, err := r.List()
	assert.Nil(t, err, "Fail to list repo files")
	assert.Equal(t, len(names), len(testData), "Fail to add or list the same number")
	for _, name := range names {
		_, ok := testData[name]
		assert.Equal(t, ok, true, "Fail to list the correct data")
	}

	// remove
	removeFile := "appA"
	err = r.Remove(removeFile)
	assert.Nil(t, err, "Fail to remove valid file")
	err = r.Remove(removeFile)
	assert.NotNil(t, err, "Fail to remove invalid file")

	// update (add with a exist name)
	updateFile := "appB"
	updateContent := "This is the content of updated appB."
	err = r.Add(updateFile, []byte(updateContent))
	assert.Nil(t, err, "Fail to add an exist file")

	// get
	_, err = r.Get(removeFile)
	assert.NotNil(t, err, "Fail to get removed file")

	res, err := r.Get(updateFile)
	assert.Nil(t, err, "Fail to get file")
	assert.Equal(t, string(res), updateContent)

	// get meta
	metas, err := r.GetMeta()
	assert.Nil(t, err, "Fail to get meta data")
	assert.Equal(t, len(metas), 1, "Fail to get correct meta count")
}
