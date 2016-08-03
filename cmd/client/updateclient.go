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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerops/dockyard/utils"
)

const (
	defaultMeta      = "meta.json"
	defaultMetaSign  = "meta.sign"
	defaultPublicKey = "pub.pem"
)

// UpdateClient reprensents the local client interface
type UpdateClient struct {
	config UpdateClientConfig
}

func (uc *UpdateClient) List(repoURL string) ([]string, error) {
	repo, err := NewUCRepo(repoURL)
	if err != nil {
		return nil, err
	}

	return repo.List()
}

func (uc *UpdateClient) GetMeta(repoURL string) (string, error) {
	repo, err := NewUCRepo(repoURL)
	if err != nil {
		return "", err
	}

	fileBytes, err := repo.GetMeta()
	if err != nil {
		return "", err
	}

	return uc.save(repo, defaultMeta, fileBytes)
}

func (uc *UpdateClient) GetMetaSign(repoURL string) (string, error) {
	repo, err := NewUCRepo(repoURL)
	if err != nil {
		return "", err
	}

	fileBytes, err := repo.GetMetaSign()
	if err != nil {
		return "", err
	}

	return uc.save(repo, defaultMetaSign, fileBytes)
}

func (uc *UpdateClient) GetPublicKey(repoURL string) (string, error) {
	repo, err := NewUCRepo(repoURL)
	if err != nil {
		return "", err
	}

	fileBytes, err := repo.GetPublicKey()
	if err != nil {
		return "", err
	}

	return uc.save(repo, defaultPublicKey, fileBytes)
}

func (uc *UpdateClient) GetFile(repoURL string, name string) (string, error) {
	repo, err := NewUCRepo(repoURL)
	if err != nil {
		return "", err
	}

	fileBytes, err := repo.GetFile(name)
	if err != nil {
		return "", err
	}

	return uc.save(repo, name, fileBytes)
}

func (uc *UpdateClient) save(repo UpdateClientRepo, file string, bytes []byte) (string, error) {
	uc.config.Load()
	localFile := filepath.Join(uc.config.GetCacheDir(), repo.NRString(), file)
	if !utils.IsDirExist(filepath.Dir(localFile)) {
		os.MkdirAll(filepath.Dir(localFile), 0755)
	}

	err := ioutil.WriteFile(localFile, bytes, 0644)
	if err != nil {
		return "", err
	}

	return localFile, nil
}
