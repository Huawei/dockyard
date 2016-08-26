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
package snapshot

import (
	"errors"
)

var (
	byContainerName   = "bycontainer"
	byContainerProtos = []string{"appv1", "dockerv1"}
)

type UpdateServiceSnapshotByContainer struct {
	ID       string
	URL      string
	ItemName string

	callback Callback
}

func init() {
	RegisterSnapshot(byContainerName, &UpdateServiceSnapshotByContainer{})
}

func (m *UpdateServiceSnapshotByContainer) New(id, url, itemname string, callback Callback) (UpdateServiceSnapshot, error) {
	if id == "" || url == "" || itemname == "" {
		return nil, errors.New("'ID' , 'URL', 'ItemName' should not be empty")
	}

	m.ID, m.URL, m.ItemName, m.callback = id, url, itemname, callback
	return m, nil
}

func (m *UpdateServiceSnapshotByContainer) Supported(proto string) bool {
	for _, p := range byContainerProtos {
		if p == proto {
			return true
		}
	}

	return false
}

func (m *UpdateServiceSnapshotByContainer) Process() error {
	return nil
}

func (m *UpdateServiceSnapshotByContainer) Description() string {
	return "Group Snapshot. Scan the package/image by container, return its output"
}
