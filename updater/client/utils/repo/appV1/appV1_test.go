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
	"testing"

	"github.com/stretchr/testify/assert"

	duc_utils "github.com/containerops/dockyard/updater/client/utils"
)

// TestInitConfig tests the Init function
func TestInitConfig(t *testing.T) {
	var appV1 DyUpdaterClientAppV1Repo

	validURL := "appV1://containerops.me/containerops/offical"
	f, err := appV1.New(validURL)
	assert.Nil(t, err, "Fail to setup a valid repo")
	assert.Equal(t, f.String(), validURL, "Fail to parse url")

	invalidURL := "appInvalid://containerops.me/containerops/offical"
	_, err = appV1.New(invalidURL)
	assert.Equal(t, err, duc_utils.ErrorsDURRepoInvalid, "Fail to parse invalid url")

	invalidURL2 := "appV1://containerops.me/containerops"
	_, err = appV1.New(invalidURL2)
	assert.Equal(t, err, duc_utils.ErrorsDURRepoInvalid, "Fail to parse invalid url")
}
