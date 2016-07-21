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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
	dy_utils "github.com/containerops/dockyard/utils"
)

const (
	localPrefix = "local"
)

var (
	// Parse "local://tmp/containerops" and get  "Path" : "/tmp/containerops"
	localRegexp = regexp.MustCompile(`^(.+):/(.+)$`)
)

type DyUpdaterServerLocal struct {
	Path string
}

func init() {
	dus_utils.RegisterStorage(localPrefix, &DyUpdaterServerLocal{})
}

func (dusl *DyUpdaterServerLocal) Supported(url string) bool {
	return strings.HasPrefix(url, localPrefix+"://")
}

func (dusl *DyUpdaterServerLocal) New(url string) (dus_utils.DyUpdaterServerStorage, error) {
	parts := localRegexp.FindStringSubmatch(url)
	if len(parts) != 3 || parts[1] != localPrefix {
		return nil, dus_utils.ErrorsDUSSInvalid
	}

	dusl.Path = parts[2]

	return dusl, nil
}

func (dusl *DyUpdaterServerLocal) String() string {
	return fmt.Sprintf("%s:/%s", localPrefix, dusl.Path)
}

// Key is "namespace/repository/appname"
func (dusl *DyUpdaterServerLocal) Get(key string) ([]byte, error) {
	if !dus_utils.ValidKey(key) {
		return nil, dus_utils.ErrorsDUSSInvalidKey
	}

	//TODO: now using data, need to hash it
	dataFileName := "data"
	file := filepath.Join(dusl.Path, key, dataFileName)
	return ioutil.ReadFile(file)
}

// Key is "namespace/repository/appname"
func (dusl *DyUpdaterServerLocal) GetMeta(key string) (meta dus_utils.Meta, err error) {
	if !dus_utils.ValidKey(key) {
		return meta, dus_utils.ErrorsDUSSInvalidKey
	}

	metaFileName := "meta.json"
	filename := filepath.Join(dusl.Path, key, metaFileName)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return dus_utils.Meta{}, err
	}
	err = json.Unmarshal(data, &meta)
	return meta, err
}

// Key is "namespace/repository/appname"
func (dusl *DyUpdaterServerLocal) Put(key string, content []byte) error {
	if !dus_utils.ValidKey(key) {
		return dus_utils.ErrorsDUSSInvalidKey
	}

	topDir := filepath.Join(dusl.Path, key)
	if !dy_utils.IsDirExist(topDir) {
		if err := os.MkdirAll(topDir, 0777); err != nil {
			return err
		}
	}

	dataFileName := "data"
	dataFile := filepath.Join(topDir, dataFileName)
	if err := ioutil.WriteFile(dataFile, content, 0644); err != nil {
		return err
	}

	metaFileName := "meta.json"
	metaFile := filepath.Join(topDir, metaFileName)
	meta := dus_utils.GenerateMeta(key, content)
	metaContent, _ := json.Marshal(meta)
	if err := ioutil.WriteFile(metaFile, metaContent, 0644); err != nil {
		os.RemoveAll(topDir)
		return err
	}
	return nil
}

// Key is "namespace/repository"
func (dusl *DyUpdaterServerLocal) List(key string) ([]string, error) {
	path := filepath.Join(dusl.Path, key)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, info := range files {
		if info.IsDir() {
			names = append(names, info.Name())
		}
	}

	return names, nil
}
