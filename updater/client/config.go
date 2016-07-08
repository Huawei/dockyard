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
	"errors"
	"os"
	"path/filepath"

	"github.com/astaxie/beego/config"

	"github.com/containerops/dockyard/utils"
)

var (
	ErrorsDUCConfigExist = errors.New("dockyard update client configuration is already exist")
)

const (
	topDir       = ".dockyard"
	configDriver = "ini"
	configName   = "config.ini"
	cacheDir     = "cache"
)

type dyUpdaterConfig struct {
	DefaultServer string
	CacheDir      string
}

func (dyc *dyUpdaterConfig) Init() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	topURL := filepath.Join(homeDir, topDir)
	if !utils.IsDirExist(topURL) {
		if err := os.Mkdir(topURL, os.ModePerm); err != nil {
			return err
		}
	}

	cacheURL := filepath.Join(topURL, cacheDir)
	if !utils.IsDirExist(cacheURL) {
		if err := os.Mkdir(cacheURL, os.ModePerm); err != nil {
			return err
		}
	}

	configFile := filepath.Join(topURL, configName)
	if utils.IsFileExist(configFile) {
		return ErrorsDUCConfigExist
	} else {
		if _, err := os.Create(configFile); err != nil {
			return err
		}
	}

	if conf, err := config.NewConfig(configDriver, configFile); err != nil {
		return err
	} else {
		conf.Set("DefaultServer", "localhost")
		conf.Set("CacheDir", cacheURL)
		if err := conf.SaveConfigFile(configFile); err != nil {
			return err
		}
	}

	return nil
}

func (dyc *dyUpdaterConfig) Load() error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return errors.New("Cannot get home directory")
	}

	conf, err := config.NewConfig(configDriver, filepath.Join(homeDir, topDir, configName))
	if err != nil {
		return err
	}

	dyc.DefaultServer = conf.String("DefaultServer")
	dyc.CacheDir = conf.String("CacheDir")
	if dyc.CacheDir == "" {
		dyc.CacheDir = filepath.Join(homeDir, topDir, cacheDir)
	}

	return nil
}
