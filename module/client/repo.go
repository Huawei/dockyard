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

package module

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// UpdateClientRepo reprensents the local repo interface
type UpdateClientRepo interface {
	Supported(url string) bool
	New(url string) (UpdateClientRepo, error)
	List() ([]string, error)
	GetFile(name string) ([]byte, error)
	GetPublicKey() ([]byte, error)
	GetMeta() ([]byte, error)
	GetMetaSign() ([]byte, error)
	Put(name string, content []byte) error
	Delete(name string) error
	NRString() string
	String() string
}

var (
	ucReposLock sync.Mutex
	ucRepos     = make(map[string]UpdateClientRepo)

	// ErrorsUCRepoInvalid occurs when a repository is invalid
	ErrorsUCRepoInvalid = errors.New("repository is invalid")
	// ErrorsUCRepoNotSupported occurs when a url is not supported by existed implementations
	ErrorsUCRepoNotSupported = errors.New("repository protocal is not supported")
)

// RegisterRepo provides a way to dynamically register an implementation of a
// Repo.
//
// If RegisterRepo is called twice with the same name if Repo is nil,
// or if the name is blank, it panics.
func RegisterRepo(name string, f UpdateClientRepo) {
	if name == "" {
		panic("Could not register a Repo with an empty name")
	}
	if f == nil {
		panic("Could not register a nil Repo")
	}

	ucReposLock.Lock()
	defer ucReposLock.Unlock()

	if _, alreadyExists := ucRepos[name]; alreadyExists {
		panic(fmt.Sprintf("Repo type '%s' is already registered", name))
	}
	ucRepos[name] = f
}

// NewUCRepo creates a update client repo by a url
func NewUCRepo(url string) (UpdateClientRepo, error) {
	//URL should be protocal#repourl
	s := strings.Split(url, "#")
	if len(s) != 2 {
		return nil, ErrorsUCRepoInvalid
	}

	for _, f := range ucRepos {
		if f.Supported(s[0]) {
			return f.New(s[1])
		}
	}

	return nil, ErrorsUCRepoNotSupported
}
