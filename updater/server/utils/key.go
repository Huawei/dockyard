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

func GenerateRSAKeyPair(bits int) ([]byte, []byte, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)

	if err != nil {
		return nil, nil, err
	}

	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}
	pubBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes}

	return pem.EncodeToMemory(privBlock), pem.EncodeToMemory(pubBlock), nil
}

func RSAEncrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, contentBytes)
}

func RSADecrypt(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privKey, contentBytes)
}

func SHA256Sign(keyBytes []byte, contentBytes []byte) ([]byte, error) {
	privKey, err := getPrivKey(keyBytes)
	if err != nil {
		return nil, err
	}

	hashed := sha256.Sum256(contentBytes)
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
}

func SHA256Verify(keyBytes []byte, contentBytes []byte, signBytes []byte) error {
	pubKey, err := getPubKey(keyBytes)
	if err != nil {
		return err
	}

	signStr := hex.EncodeToString(signBytes)
	newSignBytes, _ := hex.DecodeString(signStr)
	hashed := sha256.Sum256(contentBytes)
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], newSignBytes)
}

func getPrivKey(privBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("Fail to decode private key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func getPubKey(pubBytes []byte) (*rsa.PublicKey, error) {
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
