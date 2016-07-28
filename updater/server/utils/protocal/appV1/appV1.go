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
	dus_utils "github.com/containerops/dockyard/updater/server/utils"
	_ "github.com/containerops/dockyard/updater/server/utils/storage/local"
)

const (
	appV1Prefix = "appV1"
)

type DyUpdaterServerAppV1 struct {
}

func init() {
	dus_utils.RegisterProtocal(appV1Prefix, &DyUpdaterServerAppV1{})
}

func (ap *DyUpdaterServerAppV1) Supported(protocal string) bool {
	return protocal == appV1Prefix
}

func (ap *DyUpdaterServerAppV1) New(protocal string) (dus_utils.DyUpdaterServerProtocal, error) {
	if protocal != appV1Prefix {
		return nil, dus_utils.ErrorsDUSPNotSupported
	}
	return ap, nil
}

func (ap *DyUpdaterServerAppV1) Put(key string, data []byte) error {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return err
	}

	return s.Put(key, data)
}

func (ap *DyUpdaterServerAppV1) Get(key string) ([]byte, error) {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return nil, err
	}

	return s.Get(key)
}

func (ap *DyUpdaterServerAppV1) List(key string) ([]string, error) {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return nil, err
	}

	return s.List(key)
}

func (ap *DyUpdaterServerAppV1) GetPublicKey(key string) ([]byte, error) {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return nil, err
	}

	return s.GetPublicKey(key)
}

func (ap *DyUpdaterServerAppV1) GetMeta(key string) ([]byte, error) {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return nil, err
	}

	return s.GetMeta(key)
}

func (ap *DyUpdaterServerAppV1) GetMetaSign(key string) ([]byte, error) {
	s, err := dus_utils.NewDUSStorage("")
	if err != nil {
		return nil, err
	}

	return s.GetMetaSign(key)
}
