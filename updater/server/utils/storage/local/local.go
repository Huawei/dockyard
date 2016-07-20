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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
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

func (dusl *DyUpdaterServerLocal) Get(key string) ([]byte, error) {
	return nil, nil
}

func (dusl *DyUpdaterServerLocal) Put(key string, content []byte) error {
	return nil
}

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
