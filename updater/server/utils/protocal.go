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

type DyUpdaterServerProtocal interface {
	Supported(protocal string) bool
	New(protocal string) (DyUpdaterServerProtocal, error)
	List(key string) ([]string, error)
	GetMeta(key string) ([]Meta, error)
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
}

var (
	dusProtocalsLock sync.Mutex
	dusProtocals     = make(map[string]DyUpdaterServerProtocal)

	ErrorsDUSPNotSupported = errors.New("protocal is not supported")
	ErrorsDUSSInvalid      = errors.New("protocal url is invalid")
)

// RegisterProtocal provides a way to dynamically register an implementation of a
// protocal.
//
// If RegisterProtocal is called twice with the same name if 'protocal' is nil,
// or if the name is blank, it panics.
func RegisterProtocal(name string, f DyUpdaterServerProtocal) {
	if name == "" {
		panic("Could not register a Protocal with an empty name")
	}
	if f == nil {
		panic("Could not register a nil Protocal")
	}

	dusProtocalsLock.Lock()
	defer dusProtocalsLock.Unlock()

	if _, alreadyExists := dusProtocals[name]; alreadyExists {
		panic(fmt.Sprintf("Protocal type '%s' is already registered", name))
	}
	dusProtocals[name] = f
}

func NewDUSProtocal(protocal string) (DyUpdaterServerProtocal, error) {
	for _, f := range dusProtocals {
		if f.Supported(protocal) {
			return f.New(protocal)
		}
	}

	return nil, ErrorsDUSPNotSupported
}
