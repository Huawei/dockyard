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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
)

func readBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return ioutil.ReadAll(f)
}

func RSAEncrypt(path string, contentByte []byte) ([]byte, error) {
	pubKey, err := getPubKey(path)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, contentByte)
}

func RSADecrypt(path string, contentByte []byte) ([]byte, error) {
	privKey, err := getPrivKey(path)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privKey, contentByte)
}

func SHA256Sign(path string, contentByte []byte) ([]byte, error) {
	privKey, err := getPrivKey(path)
	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256(contentByte)
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

func SHA256Verify(path string, contentByte []byte, signByte []byte) error {
	pubKey, err := getPubKey(path)
	if err != nil {
		return err
	}

	signStr := hex.EncodeToString(signByte)
	newSignByte, _ := hex.DecodeString(signStr)
	hashed := sha256.Sum256(contentByte)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], newSignByte)
}

func getPrivKey(path string) (*rsa.PrivateKey, error) {
	privBytes, err := readBytes(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("Fail to decode private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func getPubKey(path string) (*rsa.PublicKey, error) {
	pubBytes, err := readBytes(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return nil, errors.New("Fail to decode public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Fail get public key from public interface")
	}

	return pubKey, nil
}
