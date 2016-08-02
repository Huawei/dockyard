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

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
)

var (
	urlRegexp = regexp.MustCompile(`^(.+)/(.+)/(.+)$`)
	// ErrorInvalidLocalRepo occurs when a local url is invalid
	ErrorInvalidLocalRepo = errors.New("Invalid local url")
	// ErrorEmptyRepo occurs when a repo is empty
	ErrorEmptyRepo = errors.New("Repo is empty")
	// ErrorAppNotExist occurs when an app is not exist
	ErrorAppNotExist = errors.New("App is not exist")
)

const (
	defaultMeta      = "meta.json"
	defaultMetaSign  = "meta.sig"
	defaultKeyDir    = "key"
	defaultPubKey    = "pub_key.pem"
	defaultTargetDir = "target"
)

// Repo reprensents a local repository
// if Repo is {
//	Protocal: "protocalA/VersionB",
//	Path: "/data",
//      Namespace: "containerops",
//      Repository: "official",
//      }
//    add assume there are 'osX/archY/appA' and 'osX/archY/appB' files.
// The local structure will be:
// /data
//   |_ protocalA
//       |_ VersionB
//            |_ containerops
//                |_ official
//                    |_ meta.json
//                    |_ meta.sig
//                    |_ key
//                    |    |_ pub_key.pem
//                    |
//                    |_ target
//                          |_ hashOfappA
//                          |_ hashOfappB
type Repo struct {
	Protocal   string
	Path       string
	Namespace  string
	Repository string

	kmURL string
}

// NewRepo gets repo by a protocal, a local storage path and a url
// nr : "namespace/repository"
func NewRepo(path string, protocal string, nr string) (Repo, error) {
	parts := strings.Split(nr, "/")
	if len(parts) != 2 {
		return Repo{}, ErrorInvalidLocalRepo
	}

	repo := Repo{Protocal: protocal, Path: path, Namespace: parts[0], Repository: parts[1]}

	// create top dir if not exist
	topDir := repo.GetTopDir()
	if !utils.IsDirExist(topDir) {
		if err := os.MkdirAll(topDir, 0777); err != nil {
			return Repo{}, err
		}
	}

	repo.kmURL = ""
	return repo, nil
}

// NewRepoWithKM gets repo by a protocal, a local storage path, a url and
//   a keymanager url
// nr : "namespace/repository"
// kmURL: nil means using the km repository defined in configuration
func NewRepoWithKM(path string, protocal, nr string, kmURL string) (Repo, error) {
	// if kmURL == "", try the one in setting
	if kmURL == "" {
		kmURL = setting.KeyManager
	}

	repo, err := NewRepo(path, protocal, nr)
	if err != nil {
		return Repo{}, err
	} else if kmURL == "" {
		return repo, err
	}

	err = repo.SetKM(kmURL)
	if err != nil {
		return Repo{}, err
	}

	return repo, nil
}

// SetKM sets the keymanager
func (r *Repo) SetKM(kmURL string) error {
	// pull the public key
	km, err := module.NewKeyManager(kmURL)
	if err != nil {
		return err
	}

	data, err := km.GetPublicKey(r.Protocal, r.Namespace+"/"+r.Repository)
	if err != nil {
		return err
	}

	keyfile := r.GetPublicKeyFile()
	if !utils.IsDirExist(filepath.Dir(keyfile)) {
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

// GetTopDir gets the top directory of a repository
func (r Repo) GetTopDir() string {
	return filepath.Join(r.Path, r.Protocal, r.Namespace, r.Repository)
}

// GetMetaFile gets the meta data file url of repository
func (r Repo) GetMetaFile() string {
	return filepath.Join(r.GetTopDir(), defaultMeta)
}

// GetMetaSignFile gets the meta signature file url of repository
func (r Repo) GetMetaSignFile() string {
	return filepath.Join(r.GetTopDir(), defaultMetaSign)
}

// GetPublicKeyFile gets the public key file url of repository
func (r Repo) GetPublicKeyFile() string {
	return filepath.Join(r.GetTopDir(), defaultKeyDir, defaultPubKey)
}

// GetMeta gets the meta data of a repository
func (r Repo) GetMeta() ([]byte, error) {
	metaFile := r.GetMetaFile()
	if !utils.IsFileExist(metaFile) {
		return nil, ErrorEmptyRepo
	}

	return ioutil.ReadFile(metaFile)
}

// List lists the applications inside a repository
func (r Repo) List() ([]string, error) {
	data, err := r.GetMeta()
	if err != nil {
		return nil, err
	}

	var metas []utils.Meta
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

// Get gets the data of an application
func (r Repo) Get(name string) ([]byte, error) {
	var metas []utils.Meta
	metaFile := r.GetMetaFile()
	if !utils.IsFileExist(metaFile) {
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

	for i := range metas {
		if metas[i].Name != name {
			continue
		}

		dataFile := filepath.Join(r.GetTopDir(), defaultTargetDir, metas[i].Hash)
		return ioutil.ReadFile(dataFile)
	}

	return nil, ErrorAppNotExist
}

// Put adds an application to a repository
func (r Repo) Put(name string, content []byte) (string, error) {
	topDir := r.GetTopDir()
	if !utils.IsDirExist(topDir) {
		if err := os.MkdirAll(topDir, 0777); err != nil {
			return "", err
		}
	}

	var metas []utils.Meta
	metaFile := r.GetMetaFile()
	if utils.IsFileExist(metaFile) {
		data, err := ioutil.ReadFile(metaFile)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(data, &metas)
		if err != nil {
			return "", err
		}

	}
	meta := utils.GenerateMeta(name, content)

	// Using the 'hash' value to rename the original file
	dataFileName := meta.GetHash()
	dataFile := filepath.Join(topDir, defaultTargetDir, dataFileName)
	if !utils.IsDirExist(filepath.Dir(dataFile)) {
		if err := os.MkdirAll(filepath.Dir(dataFile), 0777); err != nil {
			return "", err
		}
	}

	// write data
	if err := ioutil.WriteFile(dataFile, content, 0644); err != nil {
		return "", err
	}

	// get meta content
	exist := false
	for i := range metas {
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
		return "", err
	}

	return dataFile, nil
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

	km, _ := module.NewKeyManager(r.kmURL)
	signContent, _ := km.Sign(r.Protocal, r.Namespace+"/"+r.Repository, metasContent)
	signFile := r.GetMetaSignFile()
	if err := ioutil.WriteFile(signFile, signContent, 0644); err != nil {
		return err
	}

	return nil
}

// Remove removes an application from a repository
func (r Repo) Remove(name string) error {
	var metas []utils.Meta
	metaFile := r.GetMetaFile()
	if !utils.IsFileExist(metaFile) {
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

	for i := range metas {
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
