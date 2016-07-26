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
	"path/filepath"
	"regexp"
	"strings"

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
)

const (
	localPrefix       = "local"
	keyDir            = "key"
	defaultPublicKey  = "pub_key.pem"
	defaultPrivateKey = "priv_key.pem"
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

// Key is "namespace/repository"
func (dkml *DyKeyManagerLocal) GetPublicKey(key string) ([]byte, error) {
	file := filepath.Join(dkml.Path, key, keyDir, defaultPublicKey)

	return ioutil.ReadFile(file)
}

// Key is "namespace/repository"
func (dkml *DyKeyManagerLocal) Sign(key string, data []byte) ([]byte, error) {
	file := filepath.Join(dkml.Path, key, keyDir, defaultPrivateKey)
	return dus_utils.SHA256Sign(file, data)
}
