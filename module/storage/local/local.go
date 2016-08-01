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
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/containerops/dockyard/module"
)

const (
	localPrefix = "local"
)

var (
	// Parse "local://tmp/containerops" and get  "Path" : "/tmp/containerops"
	localRegexp = regexp.MustCompile(`^(.+):/(.+)$`)
)

// UpdateServiceStorageLocal is the local file implementation of storage service
type UpdateServiceStorageLocal struct {
	Path string

	kmURL string
}

func init() {
	module.RegisterStorage(localPrefix, &UpdateServiceStorageLocal{})
}

// Supported checks if a url begin with 'local://'
func (ussl *UpdateServiceStorageLocal) Supported(url string) bool {
	return strings.HasPrefix(url, localPrefix+"://")
}

// New creates an UpdateServceStorage interface with a local implmentation
func (ussl *UpdateServiceStorageLocal) New(url string, km string) (module.UpdateServiceStorage, error) {
	parts := localRegexp.FindStringSubmatch(url)
	if len(parts) != 3 || parts[1] != localPrefix {
		return nil, errors.New("invalid url set in StorageLocal.New")
	}

	ussl.Path = parts[2]
	ussl.kmURL = km

	return ussl, nil
}

// String returns the composed url
func (ussl *UpdateServiceStorageLocal) String() string {
	return fmt.Sprintf("%s:/%s", localPrefix, ussl.Path)
}

// Get the data of an input key. Key is "namespace/repository/appname"
func (ussl *UpdateServiceStorageLocal) Get(key string) ([]byte, error) {
	s := strings.Split(key, "/")
	if len(s) != 3 {
		return nil, errors.New("invalid key detected in StorageLocal.Get")
	}

	r, err := NewRepoWithKM(ussl.Path, strings.Join(s[:2], "/"), ussl.kmURL)
	if err != nil {
		return nil, err
	}

	return r.Get(s[2])
}

// GetMeta gets the metadata of an input key. Key is "namespace/repository"
func (ussl *UpdateServiceStorageLocal) GetMeta(key string) ([]byte, error) {
	s := strings.Split(key, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid key detected in StorageLocal.GetMeta")
	}

	r, err := NewRepoWithKM(ussl.Path, key, ussl.kmURL)
	if err != nil {
		return nil, err
	}

	return r.GetMeta()
}

// GetMetaSign gets the meta signature data. Key is "namespace/repository"
func (ussl *UpdateServiceStorageLocal) GetMetaSign(key string) ([]byte, error) {
	s := strings.Split(key, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid key detected in StorageLocal.GetMetaSign")
	}

	r, err := NewRepoWithKM(ussl.Path, key, ussl.kmURL)
	if err != nil {
		return nil, err
	}

	file := r.GetMetaSignFile()
	return ioutil.ReadFile(file)
}

// GetPublicKey gets the public key data. Key is "namespace/repository"
func (ussl *UpdateServiceStorageLocal) GetPublicKey(key string) ([]byte, error) {
	s := strings.Split(key, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid key detected in StorageLocal.GetPublicKey")
	}

	r, err := NewRepoWithKM(ussl.Path, key, ussl.kmURL)
	if err != nil {
		return nil, err
	}

	file := r.GetPublicKeyFile()
	return ioutil.ReadFile(file)
}

// Put adds a file with a key. Key is "namespace/repository/appname"
func (ussl *UpdateServiceStorageLocal) Put(key string, content []byte) error {
	s := strings.Split(key, "/")
	if len(s) != 3 {
		return errors.New("invalid key detected in StorageLocal.Put")
	}

	r, err := NewRepoWithKM(ussl.Path, strings.Join(s[:2], "/"), ussl.kmURL)
	if err != nil {
		return err
	}
	return r.Add(s[2], content)
}

// Delete removes a file by a key. Key is "namespace/repository"
func (ussl *UpdateServiceStorageLocal) Delete(key string) error {
	s := strings.Split(key, "/")
	if len(s) != 2 {
		return errors.New("invalid key detected in StorageLocal.Delete")
	}

	r, err := NewRepoWithKM(ussl.Path, strings.Join(s[:2], "/"), ussl.kmURL)
	if err != nil {
		return err
	}

	return r.Remove(s[2])
}

// List lists the content of a key. Key is "namespace/repository"
func (ussl *UpdateServiceStorageLocal) List(key string) ([]string, error) {
	s := strings.Split(key, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid key deteced in StorageLocal.List")
	}

	r, err := NewRepoWithKM(ussl.Path, key, ussl.kmURL)
	if err != nil {
		return nil, err
	}

	return r.List()
}
