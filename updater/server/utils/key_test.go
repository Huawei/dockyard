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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRSAGenerateEnDe(t *testing.T) {
	privBytes, pubBytes, err := GenerateRSAKeyPair(1024)
	assert.Nil(t, err, "Fail to genereate RSA Key Pair")

	testData := []byte("This is the testdata for encrypt and decryp")
	encrypted, err := RSAEncrypt(pubBytes, testData)
	assert.Nil(t, err, "Fail to encrypt data")
	decrypted, err := RSADecrypt(privBytes, encrypted)
	assert.Nil(t, err, "Fail to decrypt data")
	assert.Equal(t, testData, decrypted, "Fail to get correct data after en/de")
}

// TestSHA256Sign
func TestSHA256Sign(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(path), "testdata")

	testPrivFile := filepath.Join(dir, "rsa_private_key.pem")
	testContentFile := filepath.Join(dir, "hello.txt")
	testSignFile := filepath.Join(dir, "hello.sig")

	privBytes, _ := readBytes(testPrivFile)
	signBytes, _ := readBytes(testSignFile)
	contentBytes, _ := readBytes(testContentFile)
	testBytes, err := SHA256Sign(privBytes, contentBytes)
	assert.Nil(t, err, "Fail to sign")
	assert.Equal(t, testBytes, signBytes, "Fail to get valid sign data ")
}

// TestSHA256Verify
func TestSHA256Verify(t *testing.T) {
	_, path, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(path), "testdata")

	testPubFile := filepath.Join(dir, "rsa_public_key.pem")
	testContentFile := filepath.Join(dir, "hello.txt")
	testSignFile := filepath.Join(dir, "hello.sig")

	pubBytes, _ := readBytes(testPubFile)
	signBytes, _ := readBytes(testSignFile)
	contentBytes, _ := readBytes(testContentFile)
	err := SHA256Verify(pubBytes, contentBytes, signBytes)
	assert.Nil(t, err, "Fail to verify valid signed data")
	err = SHA256Verify(pubBytes, []byte("Invalid content data"), signBytes)
	assert.NotNil(t, err, "Fail to verify invalid signed data")
}
