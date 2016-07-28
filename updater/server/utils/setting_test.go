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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetting
func TestSetting(t *testing.T) {
	err := SetSetting("", "content")
	assert.NotNil(t, err, "Fail to set wrong setting")

	err = SetSetting("a", "content")
	assert.Nil(t, err, "Fail to set correct setting")

	err = SetSetting("a", "new content")
	assert.Nil(t, err, "Fail to set correct setting")

	_, err = GetSetting("")
	assert.NotNil(t, err, "Fail to get empty key")

	_, err = GetSetting("b")
	assert.NotNil(t, err, "Fail to get non exist key")

	v, err := GetSetting("a")
	assert.Nil(t, err, "Fail to get exist key")
	assert.Equal(t, v, "new content", "Fail to get correct value")
}
