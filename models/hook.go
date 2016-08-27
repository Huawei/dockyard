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

package models

import (
	"errors"
	"time"

	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/updateservice/snapshot"
	"github.com/containerops/dockyard/utils"
)

// ScanHookRegist:
//   Namespace/Repository contains images/apps/vms to be scaned
//   ScanPluginName is a plugin name of 'snapshot'
type ScanHookRegist struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	Proto          string `json:"proto" sql:"not null;type:varchar(255)"`
	Namespace      string `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository     string `json:"repository" sql:"not null;type:varchar(255)"`
	ScanPluginName string `json:"scanPluginName" sql:"not null;type:varchar(255)"`
}

// Regist regists a repository with a scan image
//   A namespace/repository could have multiple ScanPluginName,
//   but now we only support one.
func (s *ScanHookRegist) Regist(p, n, r, name string) error {
	if p == "" || n == "" || r == "" || name == "" {
		return errors.New("'Proto', 'Namespace', 'Repository' and 'ScanPluginName' should not be empty")
	}

	info := snapshot.SnapshotInputInfo{DataProto: p, Name: name}
	if ok, err := snapshot.IsSnapshotSupported(info); !ok {
		return err
	}

	s.Proto, s.Namespace, s.Repository, s.ScanPluginName = p, n, r, name
	//TODO: add to db
	return nil
}

// UnRegist unregists a repository with a scan image
//   if ImageName is nil, unregist all the scan images.
func (s *ScanHookRegist) UnRegist(p, n, r string) error {
	if p == "" || n == "" || r == "" {
		return errors.New("'Proto', 'Namespace', 'Repository' should not be empty")
	}
	s.Proto, s.Namespace, s.Repository = p, n, r

	//TODO: remove from db
	return nil
}

// FindByID finds content by id
func (s *ScanHookRegist) FindByID(id int64) (ScanHookRegist, error) {
	//TODO: query db
	return *s, nil
}

// FindID finds id by Proto, Namespace and Repository
func (s *ScanHookRegist) FindID(p, n, r string) (int64, error) {
	//TODO: query db
	return 0, nil
}

// ListScanHooks returns a list of registed scan hooks of a repository
func (s *ScanHookRegist) List(n, r string) ([]ScanHookRegist, error) {
	if n == "" || r == "" {
		return nil, errors.New("'Namespace' and 'Repository'  should not be empty")
	}
	return nil, nil
}

// ScanHookTask is the scan task
type ScanHookTask struct {
	ID int64 `json:"id" gorm:"primary_key"`
	//Path is image url now
	Path     string `json:"path" sql:"not null;type:varchar(255)"`
	Callback string `json:"callback" sql:"not null;type:varchar(255)"`
	// ID of ScanHookRegist
	RegistID int64 `json:"regist_id" sql:"not null"`
	// Status: new, running, finish
	Status    string    `json:"status" sql:"not null;type:varchar(255)"`
	Result    string    `json:"result" sql:"null;type:text"`
	CreatedAt time.Time `json:"create_at" sql:""`
	UpdatedAt time.Time `json:"update_at" sql:""`
}

// Put returns encoded id
func (t *ScanHookTask) Put(rID int64, url string) (int64, error) {
	if url == "" || rID == 0 {
		return 0, errors.New("'URL' and 'RegistID' should not be empty")
	}

	var reg ScanHookRegist
	reg, err := reg.FindByID(rID)
	if err != nil {
		return 0, err
	}

	//TODO: add to db and get task ID
	var encodedCallbackID string
	var info snapshot.SnapshotInputInfo
	info.Name = reg.ScanPluginName
	info.CallbackID = encodedCallbackID
	info.DataProto = reg.Proto
	info.DataURL = url
	// Do the real scan work
	s, err := snapshot.NewUpdateServiceSnapshot(info)
	if err != nil {
		return 0, err
	}

	err = s.Process()
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func (t *ScanHookTask) Update(status string) error {
	//TODO: update status and updatedAt
	return nil
}

func (t *ScanHookTask) Find(encodedCallbackID string) error {
	//TODO: update status and updatedAt
	var id int64
	err := utils.TokenUnmarshal(encodedCallbackID, setting.ScanKey, &id)

	return err
}

func (t *ScanHookTask) UpdateResult(encodedCallbackID string, data []byte) error {
	var id int64
	err := utils.TokenUnmarshal(encodedCallbackID, setting.ScanKey, &id)
	if err != nil {
		return err
	}
	//TODO find task from db by id

	t.Result = string(data)
	t.Update("finish")

	return nil
}

func (t *ScanHookTask) List(n, r string) ([]ScanHookTask, error) {
	if n == "" || r == "" {
		return nil, errors.New("'Namespace' and 'Repository'  should not be empty")
	}
	return nil, nil
}
