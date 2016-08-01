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
	"sync"

	"github.com/containerops/dockyard/setting"
)

// KeyManager should be seperate from dockyard
// Now only assume that keys are existed in the backend key manager.
// It is up to each implementation to decide whether provides a way
//  to generate key pair automatically.
type KeyManager interface {
	// url is the database address or local directory (local://tmp/cache)
	New(url string) (KeyManager, error)
	Supported(url string) bool
	// key : namespace/repository
	GetPublicKey(key string) ([]byte, error)
	// key : namespace/repository
	Sign(key string, data []byte) ([]byte, error)
}

var (
	kmsLock sync.Mutex
	kms     = make(map[string]KeyManager)

	// ErrorsKMNotSupported occurs when the km type is not supported
	ErrorsKMNotSupported = errors.New("key manager type is not supported")
)

// RegisterKeyManager provides a way to dynamically register an implementation of a
// storage type.
//
// If RegisterKeyManager is called twice with the same name if 'storage type' is nil,
// or if the name is blank, it panics.
func RegisterKeyManager(name string, f KeyManager) {
	if name == "" {
		panic("Could not register a KeyManager with an empty name")
	}
	if f == nil {
		panic("Could not register a nil KeyManager")
	}

	kmsLock.Lock()
	defer kmsLock.Unlock()

	if _, alreadyExists := kms[name]; alreadyExists {
		panic(fmt.Sprintf("KeyManager type '%s' is already registered", name))
	}
	kms[name] = f
}

// NewKeyManager create a key manager by a url
func NewKeyManager(url string) (KeyManager, error) {
	if url == "" {
		url = setting.KeyManager
	}
	for _, f := range kms {
		if f.Supported(url) {
			return f.New(url)
		}
	}

	return nil, ErrorsKMNotSupported
}
