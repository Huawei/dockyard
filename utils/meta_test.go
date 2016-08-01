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
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGenerate
func TestGenerate(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(path), "testdata")

	testContentFile := filepath.Join(dir, "hello.txt")
	testHashFile := filepath.Join(dir, "hello.hash")
	contentByte, _ := ioutil.ReadFile(testContentFile)
	hashByte, _ := ioutil.ReadFile(testHashFile)
	meta := GenerateMeta("hello.txt", contentByte)
	assert.Equal(t, meta.GetHash(), strings.TrimSpace(string(hashByte)), "Fail to get correct hash value")
}

// TestTime
func TestTime(t *testing.T) {
	test1 := "test1"
	test1Byte := []byte("test1 byte")
	meta1 := GenerateMeta(test1, test1Byte)
	meta2 := meta1
	assert.Equal(t, meta1, meta2, "Fail to compare meta, should be the same")

	meta2.SetCreated(meta2.GetCreated().Add(time.Hour * 1))
	cmp := meta1.Compare(meta2)
	assert.Equal(t, cmp < 0, true, "Fail to compare meta, should be smaller")

	assert.Equal(t, meta2.IsExpired(), false, "Fail to get expired information")
	meta2.SetExpired(time.Now().Add(time.Hour * (-1)))
	assert.Equal(t, meta2.IsExpired(), true, "Fail to get expired information")
}
