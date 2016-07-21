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
	"regexp"
	"sync"
)

type DyUpdaterServerStorage interface {
	// url is the database address or local directory (local://tmp/cache)
	New(url string) (DyUpdaterServerStorage, error)
	// get the 'url' set by 'New'
	String() string
	Supported(url string) bool
	Get(key string) ([]byte, error)
	GetMeta(key string) (Meta, error)
	Put(key string, data []byte) error
	List(key string) ([]string, error)
}

var (
	dusStoragesLock sync.Mutex
	dusStorages     = make(map[string]DyUpdaterServerStorage)

	keyRegexp = regexp.MustCompile(`^(.+)/(.+)/(.+)$`)

	ErrorsDUSSNotSupported = errors.New("storage type is not supported")
	ErrorsDUSSInvalidKey   = errors.New("invalid key detected")
)

// RegisterStorage provides a way to dynamically register an implementation of a
// storage type.
//
// If RegisterStorage is called twice with the same name if 'storage type' is nil,
// or if the name is blank, it panics.
func RegisterStorage(name string, f DyUpdaterServerStorage) {
	if name == "" {
		panic("Could not register a Storage with an empty name")
	}
	if f == nil {
		panic("Could not register a nil Storage")
	}

	dusStoragesLock.Lock()
	defer dusStoragesLock.Unlock()

	if _, alreadyExists := dusStorages[name]; alreadyExists {
		panic(fmt.Sprintf("Storage type '%s' is already registered", name))
	}
	dusStorages[name] = f
}

func DefaultDUSStorage() (DyUpdaterServerStorage, error) {
	//TODO: read from config
	defaultURL := "local://tmp/containerops_storage_cache"

	return NewDUSStorage(defaultURL)
}

func NewDUSStorage(url string) (DyUpdaterServerStorage, error) {
	for _, f := range dusStorages {
		if f.Supported(url) {
			return f.New(url)
		}
	}

	return nil, ErrorsDUSPNotSupported
}

func ValidStorageKey(key string) bool {
	return keyRegexp.MatchString(key)
}
