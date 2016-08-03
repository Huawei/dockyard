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
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
)

var (
	// ErrorsUCEmptyURL occurs when a repository url is nil
	ErrorsUCEmptyURL = errors.New("empty repository url")
	// ErrorsUCRepoExist occurs when a repository is exist
	ErrorsUCRepoExist = errors.New("repository is already exist")
	// ErrorsUCRepoNotExist occurs when a repository is not exist
	ErrorsUCRepoNotExist = errors.New("repository is not exist")
)

const (
	topDir     = ".dockyard"
	cacheDir   = "cache"
	configName = "repo.json"
)

// UpdateClientConfig is the local configuation of a update client
type UpdateClientConfig struct {
	DefaultServer string
	Repos         []string
}

func (ucc *UpdateClientConfig) exist() bool {
	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	return utils.IsFileExist(configFile)
}

// Init create directory and setup the cache location
func (ucc *UpdateClientConfig) Init() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	topPath := filepath.Join(homeDir, topDir)
	if !utils.IsDirExist(topPath) {
		if err := os.MkdirAll(topPath, os.ModePerm); err != nil {
			return err
		}
	}
	if !utils.IsDirExist(setting.Storage) {
		if err := os.MkdirAll(setting.Storage, os.ModePerm); err != nil {
			return err
		}
	}

	if !ucc.exist() {
		return ucc.save()
	}
	return nil
}

func (ucc *UpdateClientConfig) save() error {
	data, err := json.MarshalIndent(ucc, "", "\t")
	if err != nil {
		return err
	}

	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	if err := ioutil.WriteFile(configFile, data, 0666); err != nil {
		return err
	}

	return nil
}

// Load reads the config data
func (ucc *UpdateClientConfig) Load() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	content, err := ioutil.ReadFile(filepath.Join(homeDir, topDir, configName))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &ucc); err != nil {
		return err
	}

	return nil
}

// Add adds a repo url to the config file
func (ucc *UpdateClientConfig) Add(url string) error {
	if url == "" {
		return ErrorsUCEmptyURL
	}

	var err error
	if !ucc.exist() {
		err = ucc.Init()
	} else {
		err = ucc.Load()
	}
	if err != nil {
		return err
	}

	for _, repo := range ucc.Repos {
		if repo == url {
			return ErrorsUCRepoExist
		}
	}
	ucc.Repos = append(ucc.Repos, url)

	return ucc.save()
}

// Remove removes a repo url from the config file
func (ucc *UpdateClientConfig) Remove(url string) error {
	if url == "" {
		return ErrorsUCEmptyURL
	}

	if !ucc.exist() {
		return ErrorsUCRepoNotExist
	}

	if err := ucc.Load(); err != nil {
		return err
	}
	found := false
	for i := range ucc.Repos {
		if ucc.Repos[i] == url {
			found = true
			ucc.Repos = append(ucc.Repos[:i], ucc.Repos[i+1:]...)
			break
		}
	}
	if !found {
		return ErrorsUCRepoNotExist
	}

	return ucc.save()
}

func (ucc *UpdateClientConfig) GetCacheDir() string {
	return filepath.Join(os.Getenv("HOME"), topDir, cacheDir)
}
