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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
)

const (
	LocalPrefix       = "local"
	defaultKeyDirName = "key"
	defaultPublicKey  = "pub_key.pem"
	defaultPrivateKey = "priv_key.pem"
	defaultBitsSize   = 2048
)

var (
	// Parse "local://tmp/containerops" and get  "Path" : "/tmp/containerops"
	localRegexp = regexp.MustCompile(`^local:/(.+)$`)
)

// KeyManagerLocal is the local implementation of a key manager

type KeyManagerLocal struct {
	Path string
}

func init() {
	module.RegisterKeyManager(LocalPrefix, &KeyManagerLocal{})
}

// Supported checks if a local url begin with "local://"
func (kml *KeyManagerLocal) Supported(url string) bool {
	return strings.HasPrefix(url, LocalPrefix+"://")
}

// New returns a keymanager by a url and a protocal
func (kml *KeyManagerLocal) New(url string) (module.KeyManager, error) {
	parts := localRegexp.FindStringSubmatch(url)
	if len(parts) != 2 {
		return nil, errors.New("Invalid key manager url, should be 'local://@dir'.")
	}

	kml.Path = parts[1]
	return kml, nil
}

// GetPublicKey gets the public key data of a namespace/repository
func (kml *KeyManagerLocal) GetPublicKey(protocal string, nr string) ([]byte, error) {
	keyDir := filepath.Join(kml.Path, protocal, nr, defaultKeyDirName)
	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return nil, err
		}
	}

	return ioutil.ReadFile(filepath.Join(keyDir, defaultPublicKey))
}

// Sign signs a data of a namespace/repository
func (kml *KeyManagerLocal) Decrypt(protocal string, nr string, data []byte) ([]byte, error) {
	keyDir := filepath.Join(kml.Path, protocal, nr, defaultKeyDirName)
	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return nil, err
		}
	}

	privBytes, _ := ioutil.ReadFile(filepath.Join(keyDir, defaultPrivateKey))
	return utils.RSADecrypt(privBytes, data)
}

// Sign signs a data of a namespace/repository
func (kml *KeyManagerLocal) Sign(protocal string, nr string, data []byte) ([]byte, error) {
	keyDir := filepath.Join(kml.Path, protocal, nr, defaultKeyDirName)
	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return nil, err
		}
	}

	privBytes, _ := ioutil.ReadFile(filepath.Join(keyDir, defaultPrivateKey))
	return utils.SHA256Sign(privBytes, data)
}

func isKeyExist(keyDir string) bool {
	if !utils.IsFileExist(filepath.Join(keyDir, defaultPrivateKey)) {
		return false
	}

	if !utils.IsFileExist(filepath.Join(keyDir, defaultPublicKey)) {
		return false
	}

	return true
}

func generateKey(keyDir string) error {
	privBytes, pubBytes, err := utils.GenerateRSAKeyPair(defaultBitsSize)
	if err != nil {
		return err
	}

	if !utils.IsDirExist(keyDir) {
		err := os.MkdirAll(keyDir, 0777)
		if err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filepath.Join(keyDir, defaultPrivateKey), privBytes, 0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(keyDir, defaultPublicKey), pubBytes, 0644); err != nil {
		return err
	}

	return nil
}
