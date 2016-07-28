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
	"sync"
)

var (
	dusSettingsLock sync.Mutex
	dusSettings     = make(map[string]string)
)

func SetSetting(key string, value string) error {
	if key == "" {
		return errors.New("setting key should not be empty")
	}

	dusSettingsLock.Lock()
	defer dusSettingsLock.Unlock()

	dusSettings[key] = value
	return nil
}

func GetSetting(key string) (string, error) {
	if key == "" {
		return "", errors.New("setting key should not be empty")
	}

	if v, ok := dusSettings[key]; ok {
		return v, nil
	} else {
		return "", errors.New("setting key is not exist")
	}
}
