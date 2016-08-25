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
	"fmt"
	"sync"
)

// TODO:
type UpdateServiceSnapshotOutput struct {
	Data  []byte
	Error error
}

// Callback is a function that a snapshot plugin use after finish the `Process`
// configuration.
type Callback func(id string, output UpdateServiceSnapshotOutput) error

// UpdateServiceSnapshot represents the snapshot interface
type UpdateServiceSnapshot interface {
	// `id` : callback id
	// `url`: local file or local dir
	// `callback`: if callback is nil, the caller could handle it by itself
	//		or the caller must implement calling this in `Process`
	//		TODO: we need to certify plugins..
	New(id, url string, callback Callback) (UpdateServiceSnapshot, error)
	// `proto`: `appv1/dockerv1` for example
	Supported(proto string) bool
	Description() string
	Process() error
}

var (
	usSnapshotsLock sync.Mutex
	usSnapshots     = make(map[string]UpdateServiceSnapshot)
)

// RegisterSnapshot provides a way to dynamically register an implementation of a
// snapshot type.
func RegisterSnapshot(name string, f UpdateServiceSnapshot) error {
	if name == "" {
		return errors.New("Could not register a Snapshot with an empty name")
	}
	if f == nil {
		return errors.New("Could not register a nil Snapshot")
	}

	usSnapshotsLock.Lock()
	defer usSnapshotsLock.Unlock()

	if _, alreadyExists := usSnapshots[name]; alreadyExists {
		return fmt.Errorf("Snapshot type '%s' is already registered", name)
	}
	usSnapshots[name] = f
	return nil
}

func UnregisterAllSnapshot() {
	usSnapshotsLock.Lock()
	defer usSnapshotsLock.Unlock()

	for n, _ := range usSnapshots {
		delete(usSnapshots, n)
	}
}

func IsSnapshotSupported(proto, name string) (bool, error) {
	f, ok := usSnapshots[name]
	if !ok {
		return false, fmt.Errorf("Cannot find plugin :%s", name)
	}

	ok = f.Supported(proto)
	if !ok {
		return false, fmt.Errorf("Proto %s is not supported by plugin %s", proto, name)
	}

	return true, nil
}

func ListSnapshotByProto(proto string) (snapshots []string) {
	for n, f := range usSnapshots {
		if f.Supported(proto) {
			snapshots = append(snapshots, n)
		}
	}

	return
}

// NewUpdateServiceSnapshot creates a snapshot interface by a name and a url
func NewUpdateServiceSnapshot(name, id, url string, cb Callback) (UpdateServiceSnapshot, error) {
	f, ok := usSnapshots[name]
	if !ok {
		return nil, fmt.Errorf("Snapshot '%s' not found", name)
	}
	return f.New(id, url, cb)
}
