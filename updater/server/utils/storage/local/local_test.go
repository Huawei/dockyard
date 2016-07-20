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
	"fmt"
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
	apps, _ := l.List(key)
	fmt.Println(apps)
}
