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
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/containerops/dockyard/module/client"
	"github.com/containerops/dockyard/utils"
)

const (
	defaultMeta      = "meta.json"
	defaultMetaSign  = "meta.sign"
	defaultPublicKey = "key/pub.pem"
)

// UpdateClient reprensents the local client interface
type UpdateClient struct {
	cacheDir string
}

// Update the meta data, meta sign, public key
func (uc *UpdateClient) Update(repoURL string) error {
	repo, err := module.NewUCRepo(repoURL)
	if err != nil {
		return err
	}

	needUpdate := false
	remoteMetaBytes, err := repo.GetMeta()
	if err != nil {
		return err
	}

	localMetaFile := filepath.Join(uc.getCacheDir(), repo.NRString(), defaultMeta)
	localMetaBytes, err := ioutil.ReadFile(localMetaFile)
	if err != nil {
		needUpdate = true
	} else {
		var remoteMeta utils.Meta
		var localMeta utils.Meta
		json.Unmarshal(remoteMetaBytes, &remoteMeta)
		json.Unmarshal(localMetaBytes, &localMeta)
		needUpdate = localMeta.Before(remoteMeta)
	}

	if !needUpdate {
		fmt.Println("Already updated to the latest version.")
		return nil
	}

	uc.save(repo, defaultMeta, remoteMetaBytes)
	signBytes, _ := repo.GetMetaSign()
	pubBytes, _ := repo.GetPublicKey()
	err = utils.SHA256Verify(pubBytes, remoteMetaBytes, signBytes)
	if err != nil {
		fmt.Println("Fail to verify meta by public key")
		return err
	}

	metaURL, err := uc.save(repo, defaultMeta, remoteMetaBytes)
	if err != nil {
		return err
	}

	signURL, err := uc.save(repo, defaultMetaSign, signBytes)
	if err != nil {
		os.Remove(metaURL)
		return err
	}

	_, err = uc.save(repo, defaultPublicKey, pubBytes)
	if err != nil {
		os.Remove(metaURL)
		os.Remove(signURL)
		return err
	}

	return nil
}

// List always tells the latest apps.
func (uc *UpdateClient) List(repoURL string) ([]string, error) {
	repo, err := module.NewUCRepo(repoURL)
	if err != nil {
		return nil, err
	}

	return repo.List()
}

func (uc *UpdateClient) GetFile(repoURL string, name string) (string, error) {
	repo, err := module.NewUCRepo(repoURL)
	if err != nil {
		return "", err
	}

	fileBytes, err := repo.GetFile(name)
	if err != nil {
		return "", err
	}

	fileURL, err := uc.save(repo, name, fileBytes)
	if err != nil {
		return "", err
	}

	err = uc.Update(repoURL)
	if err != nil {
		fmt.Println("Fail to verify if downloaded file is valid")
		return fileURL, err
	}

	var meta utils.Meta
	metaBytes, _ := ioutil.ReadFile(filepath.Join(uc.cacheDir, repo.NRString(), defaultMeta))
	fileHash := fmt.Sprintf("%x", sha1.Sum(fileBytes))
	json.Unmarshal(metaBytes, &meta)
	for _, m := range meta.Items {
		if m.Name != name {
			continue
		}

		if m.Hash == fileHash {
			fmt.Println("Congratulations! The file is valid!")
			return fileURL, nil
		}

		err := errors.New("the file is invalid, maybe security issue")
		fmt.Println(err)
		return fileURL, err
	}

	return fileURL, errors.New("Cannot find the file in meta data, maybe remote system error")
}

func (uc *UpdateClient) Delete(repoURL string, name string) error {
	repo, err := module.NewUCRepo(repoURL)
	if err != nil {
		return err
	}

	return repo.Delete(name)
}

func (uc *UpdateClient) getCacheDir() string {
	if uc.cacheDir == "" {
		var config UpdateClientConfig
		config.Load()
		uc.cacheDir = config.GetCacheDir()
	}

	return uc.cacheDir
}

func (uc *UpdateClient) save(repo module.UpdateClientRepo, file string, bytes []byte) (string, error) {
	localFile := filepath.Join(uc.getCacheDir(), repo.NRString(), file)
	if !utils.IsDirExist(filepath.Dir(localFile)) {
		os.MkdirAll(filepath.Dir(localFile), 0755)
	}

	err := ioutil.WriteFile(localFile, bytes, 0644)
	if err != nil {
		return "", err
	}

	return localFile, nil
}
