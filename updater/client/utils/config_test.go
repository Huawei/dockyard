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
package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTmpHome(t *testing.T) (string, string) {
	tmpHome, err := ioutil.TempDir("", "duc-test-")
	assert.Nil(t, err, "Fail to create temp directory")

	savedHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)

	return tmpHome, savedHome
}

// TestInitConfig tests the Init function
func TestInitConfig(t *testing.T) {
	tmpHome, savedHome := createTmpHome(t)
	defer os.RemoveAll(tmpHome)
	defer os.Setenv("HOME", savedHome)

	var conf DyUpdaterClientConfig
	err := conf.Init()
	assert.Nil(t, err, "Fail to init config")
	err = conf.Init()
	assert.Equal(t, err, ErrorsDUCConfigExist, "Should not init more than once")
}

// TestLoadConfig tests the testdata/home/.dockyard/config.json file
func TestLoadConfig(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	testHome := "/testdata/home"
	savedHome := os.Getenv("HOME")
	defer os.Setenv("HOME", savedHome)
	os.Setenv("HOME", filepath.Join(filepath.Dir(path), testHome))

	var conf DyUpdaterClientConfig
	err := conf.Load()
	assert.Nil(t, err, "Fail to load config")
	assert.Equal(t, conf.DefaultServer, "containerops.me", "Fail to load 'DefaultServer'")
	assert.Equal(t, conf.CacheDir, "/tmp/containeropsCache", "Fail to load 'CacheDir'")
}

func TestAddRemoveConfig(t *testing.T) {
	tmpHome, savedHome := createTmpHome(t)
	defer os.RemoveAll(tmpHome)
	defer os.Setenv("HOME", savedHome)

	var conf DyUpdaterClientConfig
	invalidURL := ""
	validURL := "app://containerops/official/duc.rpm"

	// 'add'
	err := conf.Add(invalidURL)
	assert.Equal(t, err, ErrorsDUCInvalidRepo)
	err = conf.Add(validURL)
	assert.Nil(t, err, "Failed to add repository")
	err = conf.Add(validURL)
	assert.Equal(t, err, ErrorsDUCRepoExist)

	// 'remove'
	err = conf.Remove(validURL)
	assert.Nil(t, err, "Failed to remove repository")
	err = conf.Remove(validURL)
	assert.Equal(t, err, ErrorsDUCRepoNotExist)
}
