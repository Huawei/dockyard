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

package appV1

import (
	"github.com/containerops/dockyard/module"
)

const (
	appV1Prefix   = "appV1"
	appV1Protocal = "app/v1"
)

// UpdateServiceAppV1 is the appV1 implementation of the update service protocal
type UpdateServiceAppV1 struct {
	storage module.UpdateServiceStorage
}

func init() {
	module.Register(appV1Prefix, &UpdateServiceAppV1{})
}

// Supported checks if a protocal is 'appV1'
func (app *UpdateServiceAppV1) Supported(protocal string) bool {
	return protocal == appV1Prefix
}

// New creates a update service interface by an appV1 protocal
func (app *UpdateServiceAppV1) New(protocal string, storageURL string, kmURL string) (module.UpdateService, error) {
	if protocal != appV1Prefix {
		return nil, module.ErrorsUSPNotSupported
	}

	var err error
	app.storage, err = module.NewUpdateServiceStorage(storageURL, kmURL)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// Put adds a appV1 file to a repository
func (app *UpdateServiceAppV1) Put(nr, name string, data []byte) (string, error) {
	key := nr + "/" + name
	return app.storage.Put(appV1Protocal, key, data)
}

// Delete removes a appV1 file from a repository
func (app *UpdateServiceAppV1) Delete(nr, name string) error {
	key := nr + "/" + name
	return app.storage.Delete(appV1Protocal, key)
}

// Get gets the appV1 file data of a repository
func (app *UpdateServiceAppV1) Get(nr, name string) ([]byte, error) {
	key := nr + "/" + name
	return app.storage.Get(appV1Protocal, key)
}

// List lists the applications of a repository
func (app *UpdateServiceAppV1) List(nr string) ([]string, error) {
	return app.storage.List(appV1Protocal, nr)
}

// GetPublicKey returns the public key data of a repository
func (app *UpdateServiceAppV1) GetPublicKey(nr string) ([]byte, error) {
	return app.storage.GetPublicKey(appV1Protocal, nr)
}

// GetMeta returns the meta data of a repository
func (app *UpdateServiceAppV1) GetMeta(nr string) ([]byte, error) {
	return app.storage.GetMeta(appV1Protocal, nr)
}

// GetMetaSign returns the meta signature data of a repository
func (app *UpdateServiceAppV1) GetMetaSign(nr string) ([]byte, error) {
	return app.storage.GetMetaSign(appV1Protocal, nr)
}
