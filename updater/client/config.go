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

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerops/dockyard/utils"
)

var (
	ErrorsDUCConfigExist  = errors.New("dockyard update client configuration is already exist")
	ErrorsDUCInvalidRepo  = errors.New("invalid repository url")
	ErrorsDUCRepoExist    = errors.New("repository is already exist")
	ErrorsDUCRepoNotExist = errors.New("repository is not exist")
)

const (
	topDir     = ".dockyard"
	configName = "config.json"
	cacheDir   = "cache"
)

type dyUpdaterConfig struct {
	DefaultServer string
	CacheDir      string
	Repos         []string
}

func (dyc *dyUpdaterConfig) exist() bool {
	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	return utils.IsFileExist(configFile)
}

func (dyc *dyUpdaterConfig) Init() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	if dyc.exist() {
		return ErrorsDUCConfigExist
	}

	topURL := filepath.Join(homeDir, topDir)
	cacheURL := filepath.Join(topURL, cacheDir)
	if !utils.IsDirExist(cacheURL) {
		if err := os.MkdirAll(cacheURL, os.ModePerm); err != nil {
			return err
		}
	}

	dyc.CacheDir = cacheURL

	return dyc.save()
}

func (dyc *dyUpdaterConfig) save() error {
	data, err := json.MarshalIndent(dyc, "", "\t")
	if err != nil {
		return err
	}

	configFile := filepath.Join(os.Getenv("HOME"), topDir, configName)
	if err := ioutil.WriteFile(configFile, data, 0666); err != nil {
		return err
	}

	return nil
}

func (dyc *dyUpdaterConfig) Load() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	content, err := ioutil.ReadFile(filepath.Join(homeDir, topDir, configName))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &dyc); err != nil {
		return err
	}

	if dyc.CacheDir == "" {
		dyc.CacheDir = filepath.Join(homeDir, topDir, cacheDir)
	}

	return nil
}

func IsValidRepoURL(url string) bool {
	if url == "" {
		return false
	}

	return true
}

func (dyc *dyUpdaterConfig) Add(url string) error {
	if !IsValidRepoURL(url) {
		return ErrorsDUCInvalidRepo
	}

	var err error
	if !dyc.exist() {
		err = dyc.Init()
	} else {
		err = dyc.Load()
	}
	if err != nil {
		return err
	}

	for _, repo := range dyc.Repos {
		if repo == url {
			return ErrorsDUCRepoExist
		}
	}
	dyc.Repos = append(dyc.Repos, url)

	return dyc.save()
}

func (dyc *dyUpdaterConfig) Remove(url string) error {
	if !IsValidRepoURL(url) {
		return ErrorsDUCInvalidRepo
	}

	if !dyc.exist() {
		return ErrorsDUCRepoNotExist
	}

	if err := dyc.Load(); err != nil {
		return err
	}
	found := false
	for i, _ := range dyc.Repos {
		if dyc.Repos[i] == url {
			found = true
			dyc.Repos = append(dyc.Repos[:i], dyc.Repos[i+1:]...)
			break
		}
	}
	if !found {
		return ErrorsDUCRepoNotExist
	}

	return dyc.save()
}
