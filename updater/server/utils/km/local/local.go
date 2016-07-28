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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
	dy_utils "github.com/containerops/dockyard/utils"
)

const (
	localPrefix       = "local"
	defaultKeyDirName = "key"
	defaultPublicKey  = "pub_key.pem"
	defaultPrivateKey = "priv_key.pem"
	defaultBitsSize   = 2048
)

var (
	// Parse "local://tmp/containerops" and get  "Path" : "/tmp/containerops"
	localRegexp = regexp.MustCompile(`^(.+):/(.+)$`)
)

type DyKeyManagerLocal struct {
	Path string
}

func init() {
	dus_utils.RegisterKeyManager(localPrefix, &DyKeyManagerLocal{})
}

func (dkml *DyKeyManagerLocal) Supported(url string) bool {
	return strings.HasPrefix(url, localPrefix+"://")
}

func (dkml *DyKeyManagerLocal) New(url string) (dus_utils.DyKeyManager, error) {
	parts := localRegexp.FindStringSubmatch(url)
	if len(parts) != 3 || parts[1] != localPrefix {
		return nil, dus_utils.ErrorsDUInvalidURL
	}

	dkml.Path = parts[2]
	return dkml, nil
}

func isKeyExist(keyDir string) bool {
	if !dy_utils.IsFileExist(filepath.Join(keyDir, defaultPrivateKey)) {
		return false
	}

	if !dy_utils.IsFileExist(filepath.Join(keyDir, defaultPublicKey)) {
		return false
	}

	return true
}

func generateKey(keyDir string) error {
	privBytes, pubBytes, err := dus_utils.GenerateRSAKeyPair(defaultBitsSize)
	if err != nil {
		return err
	}

	if !dy_utils.IsDirExist(keyDir) {
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

// Key is "namespace/repository"
func (dkml *DyKeyManagerLocal) GetPublicKey(key string) ([]byte, error) {
	keyDir := filepath.Join(dkml.Path, key, defaultKeyDirName)
	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return nil, err
		}
	}

	return ioutil.ReadFile(filepath.Join(keyDir, defaultPublicKey))
}

// Key is "namespace/repository"
func (dkml *DyKeyManagerLocal) Sign(key string, data []byte) ([]byte, error) {
	keyDir := filepath.Join(dkml.Path, key, defaultKeyDirName)
	if !isKeyExist(keyDir) {
		err := generateKey(keyDir)
		if err != nil {
			return nil, err
		}
	}

	privBytes, _ := ioutil.ReadFile(filepath.Join(keyDir, defaultPrivateKey))
	return dus_utils.SHA256Sign(privBytes, data)
}
