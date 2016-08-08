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
)

// ScanHookRegist:
//   Namespace/Repository contains images/apps/vms to be scaned
//   ImageName is used to scan the images/apps/vms within a repository
//   NOTE: no need to record the type of a repository ("docker/v1", "app/v1"),
//         a user should regist his/her repository with the right ImageName
type ScanHookRegist struct {
	ID         int64  `json:"id" gorm:"primary_key"`
	Namespace  string `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository string `json:"repository" sql:"not null;type:varchar(255)"`
	ImageName  string `json:"ImageName" sql:"not null;type:varchar(255)"`
}

// Regist regists a repository with a scan image
//   A namespace/repository could have multiple ImageNames,
//   but should not have multiple records with same Namespace&Repository&ImageName.
func (s *ScanHookRegist) Regist(n, r, image string) error {
	if n == "" || r == "" || image == "" {
		return errors.New("'Namespace', 'Repository' and 'ImageName' should not be empty")
	}
	s.Namespace, s.Repository, s.ImageName = n, r, image
	//TODO: add to db
	return nil
}

// UnRegist unregists a repository with a scan image
//   if ImageName is nil, unregist all the scan images.
func (s *ScanHookRegist) UnRegist(n, r string) error {
	if n == "" || r == "" {
		return errors.New("'Namespace', 'Repository' should not be empty")
	}
	s.Namespace, s.Repository = n, r

	//TODO: remove from db
	return nil
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
	ID       int64  `json:"id" gorm:"primary_key"`
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

func (t *ScanHookTask) Put(p, c string, rID int64) error {
	if p == "" || c == "" || rID == 0 {
		return errors.New("'Namespace', 'Repository' and RegistID should not be empty")
	}

	return nil
}

func (t *ScanHookTask) Update(status string) error {
	//TODO: update status and updatedAt
	return nil
}

func (t *ScanHookTask) List(n, r string) ([]ScanHookTask, error) {
	if n == "" || r == "" {
		return nil, errors.New("'Namespace' and 'Repository'  should not be empty")
	}
	return nil, nil
}
