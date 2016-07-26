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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	dus_utils "github.com/containerops/dockyard/updater/server/utils"
	dy_utils "github.com/containerops/dockyard/utils"
)

var (
	urlRegexp             = regexp.MustCompile(`^(.+)/(.+)/(.+)$`)
	ErrorInvalidLocalRepo = errors.New("Invalid local url")
	ErrorEmptyRepo        = errors.New("Repo is empty")
	ErrorAppNotExist      = errors.New("App is not exist")
)

const (
	defaultMeta      = "meta.json"
	defaultMetaSign  = "meta.sig"
	defaultPubKeyDir = "key"
	defaultPubKey    = "pub_key.pem"
	defaultTargetDir = "target"
)

type Repo struct {
	Path       string
	Namespace  string
	Repository string

	kmURL string
}

// url : "namespace/repository"
// kmURL: nil means using the km repository defined in configuration
func NewRepo(path string, url string) (Repo, error) {
	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return Repo{}, ErrorInvalidLocalRepo
	}

	repo := Repo{Path: path, Namespace: parts[0], Repository: parts[1]}

	// create top dir if not exist
	topDir := repo.GetTopDir()
	if !dy_utils.IsDirExist(topDir) {
		if err := os.MkdirAll(topDir, 0777); err != nil {
			return Repo{}, err
		}
	}

	repo.kmURL = ""
	return repo, nil
}

func NewRepoWithKM(path string, url string, kmURL string) (Repo, error) {
	if kmURL == "" {
		return NewRepo(path, url)
	}

	repo, err := NewRepo(path, url)
	if err == nil {
		err = repo.SetKM(kmURL)
	}

	if err != nil {
		return Repo{}, err
	}

	return repo, nil
}

func (r *Repo) SetKM(kmURL string) error {
	// pull the public key
	km, err := dus_utils.NewKeyManager(kmURL)
	if err != nil {
		return err
	}

	data, err := km.GetPublicKey(r.Namespace + "/" + r.Repository)
	if err != nil {
		return err
	}

	keyfile := r.GetPublicKeyFile()
	if !dy_utils.IsDirExist(filepath.Dir(keyfile)) {
		if err := os.MkdirAll(filepath.Dir(keyfile), 0777); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(keyfile, data, 0644); err != nil {
		return err
	}

	r.kmURL = kmURL
	return nil
}

func (r Repo) GetTopDir() string {
	return filepath.Join(r.Path, r.Namespace, r.Repository)
}

func (r Repo) GetMetaFile() string {
	return filepath.Join(r.Path, r.Namespace, r.Repository, defaultMeta)
}

func (r Repo) GetMetaSignFile() string {
	return filepath.Join(r.Path, r.Namespace, r.Repository, defaultMetaSign)
}

func (r Repo) GetPublicKeyFile() string {
	return filepath.Join(r.Path, r.Namespace, r.Repository, defaultPubKeyDir, defaultPubKey)
}

func (r Repo) GetMeta() ([]dus_utils.Meta, error) {
	var metas []dus_utils.Meta
	metaFile := r.GetMetaFile()
	if !dy_utils.IsFileExist(metaFile) {
		return nil, ErrorEmptyRepo
	}

	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &metas)
	return metas, err
}

func (r Repo) List() ([]string, error) {
	metaFile := r.GetMetaFile()
	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return nil, err
	}

	var metas []dus_utils.Meta
	err = json.Unmarshal(data, &metas)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, meta := range metas {
		files = append(files, meta.Name)
	}

	return files, nil
}

func (r Repo) Get(name string) ([]byte, error) {
	var metas []dus_utils.Meta
	metaFile := r.GetMetaFile()
	if !dy_utils.IsFileExist(metaFile) {
		return nil, ErrorEmptyRepo
	}
	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &metas)
	if err != nil {
		return nil, err
	}

	for i, _ := range metas {
		if metas[i].Name != name {
			continue
		}

		dataFile := filepath.Join(r.GetTopDir(), defaultTargetDir, metas[i].Hash)
		return ioutil.ReadFile(dataFile)
	}

	return nil, ErrorAppNotExist
}

func (r Repo) Add(name string, content []byte) error {
	topDir := r.GetTopDir()
	if !dy_utils.IsDirExist(topDir) {
		if err := os.MkdirAll(topDir, 0777); err != nil {
			return err
		}
	}

	var metas []dus_utils.Meta
	metaFile := r.GetMetaFile()
	if dy_utils.IsFileExist(metaFile) {
		data, err := ioutil.ReadFile(metaFile)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &metas)
		if err != nil {
			return err
		}

	}
	meta := dus_utils.GenerateMeta(name, content)

	// Using the 'hash' value to rename the original file
	dataFileName := meta.GetHash()
	dataFile := filepath.Join(topDir, defaultTargetDir, dataFileName)
	if !dy_utils.IsDirExist(filepath.Dir(dataFile)) {
		if err := os.MkdirAll(filepath.Dir(dataFile), 0777); err != nil {
			return err
		}
	}

	// write data
	if err := ioutil.WriteFile(dataFile, content, 0644); err != nil {
		return err
	}

	// get meta content
	exist := false
	for i, _ := range metas {
		if metas[i].Name == name {
			metas[i] = meta
			exist = true
		}
	}
	if !exist {
		metas = append(metas, meta)
	}
	metasContent, _ := json.Marshal(metas)

	// write meta data
	err := r.saveMeta(metasContent)
	if err != nil {
		os.Remove(dataFile)
		return err
	}

	return nil
}

func (r Repo) saveMeta(metasContent []byte) error {
	metaFile := r.GetMetaFile()
	err := ioutil.WriteFile(metaFile, metasContent, 0644)
	if err != nil {
		return err
	}

	// write sign file
	err = r.saveSign(metasContent)
	if err != nil {
		os.Remove(metaFile)
		return err
	}

	return nil
}

func (r Repo) saveSign(metasContent []byte) error {
	if r.kmURL == "" {
		return nil
	}

	km, _ := dus_utils.NewKeyManager(r.kmURL)
	signContent, _ := km.Sign(r.Namespace+"/"+r.Repository, metasContent)
	signFile := r.GetMetaSignFile()
	if err := ioutil.WriteFile(signFile, signContent, 0644); err != nil {
		return err
	}

	return nil
}

func (r Repo) Remove(name string) error {
	var metas []dus_utils.Meta
	metaFile := r.GetMetaFile()
	if !dy_utils.IsFileExist(metaFile) {
		return ErrorEmptyRepo
	}
	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &metas)
	if err != nil {
		return err
	}

	for i, _ := range metas {
		if metas[i].Name != name {
			continue
		}

		dataFile := filepath.Join(r.GetTopDir(), defaultTargetDir, metas[i].Hash)
		if err := os.Remove(dataFile); err != nil {
			return err
		}

		metas = append(metas[:i], metas[i+1:]...)
		metasContent, _ := json.Marshal(metas)

		if err := r.saveMeta(metasContent); err != nil {
			return err
		}
		return nil
	}

	return ErrorAppNotExist
}
