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
package unittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	cutils "github.com/containerops/dockyard/cmd/client"
)

func createCCCTmpHome(t *testing.T) (string, string) {
	tmpHome, err := ioutil.TempDir("", "duc-test-")
	assert.Nil(t, err, "Fail to create temp directory")

	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)

	return tmpHome, savedHome
}

// TestCCCInitConfig tests the Init function
func TestCCCInitConfig(t *testing.T) {
	tmpHome, savedHome := createCCCTmpHome(t)
	defer os.RemoveAll(tmpHome)
	defer os.Setenv("HOME", savedHome)

	var conf cutils.UpdateClientConfig
	err := conf.Init()
	assert.Nil(t, err, "Fail to init config")
}

// TestCCCLoadConfig tests the testdata/home/.dockyard/repo.json file
func TestCCCLoadConfig(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	testHome := "/testdata/home"
	savedHome := os.Getenv("HOME")
	defer os.Setenv("HOME", savedHome)
	os.Setenv("HOME", filepath.Join(filepath.Dir(path), testHome))

	var conf cutils.UpdateClientConfig
	err := conf.Load()
	assert.Nil(t, err, "Fail to load config")
	assert.Equal(t, conf.DefaultServer, "containerops.me", "Fail to load 'DefaultServer'")
}

func TestCCCAddRemoveConfig(t *testing.T) {
	tmpHome, savedHome := createCCCTmpHome(t)
	defer os.RemoveAll(tmpHome)
	defer os.Setenv("HOME", savedHome)

	var conf cutils.UpdateClientConfig
	invalidURL := ""
	validURL := "app://containerops/official/duc.rpm"

	// 'add'
	err := conf.Add(invalidURL)
	assert.Equal(t, err, cutils.ErrorsUCEmptyURL)
	err = conf.Add(validURL)
	assert.Nil(t, err, "Failed to add repository")
	err = conf.Add(validURL)
	assert.Equal(t, err, cutils.ErrorsUCRepoExist)

	// 'remove'
	err = conf.Remove(validURL)
	assert.Nil(t, err, "Failed to remove repository")
	err = conf.Remove(validURL)
	assert.Equal(t, err, cutils.ErrorsUCRepoNotExist)
}
