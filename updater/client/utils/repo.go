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
	"errors"
	"fmt"
	"sync"
)

type DyUpdaterClientRepo interface {
	Supported(url string) bool
	New(url string) (DyUpdaterClientRepo, error)
	List() ([]string, error)
	Get() error
	String() string
}

var (
	ducReposLock sync.Mutex
	ducRepos     = make(map[string]DyUpdaterClientRepo)

	ErrorsDURRepoInvalid      = errors.New("repository is invalid")
	ErrorsDURRepoNotSupported = errors.New("repository protocal is not supported")
)

// RegisterRepo provides a way to dynamically register an implementation of a
// Repo.
//
// If RegisterRepo is called twice with the same name if Repo is nil,
// or if the name is blank, it panics.
func RegisterRepo(name string, f DyUpdaterClientRepo) {
	if name == "" {
		panic("Could not register a Repo with an empty name")
	}
	if f == nil {
		panic("Could not register a nil Repo")
	}

	ducReposLock.Lock()
	defer ducReposLock.Unlock()

	if _, alreadyExists := ducRepos[name]; alreadyExists {
		panic(fmt.Sprintf("Repo type '%s' is already registered", name))
	}
	ducRepos[name] = f
}

func NewDUCRepo(url string) (DyUpdaterClientRepo, error) {
	for _, f := range ducRepos {
		if f.Supported(url) {
			return f.New(url)
		}
	}

	return nil, ErrorsDURRepoNotSupported
}
