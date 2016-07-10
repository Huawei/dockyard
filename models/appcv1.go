/*
Copyright 2015 The ContainerOps Authors All rights reserved.

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
	"time"
)

//
type AppcV1 struct {
	Id          int64      `json:"id" gorm:"primary_key"`
	Namespace   string     `json:"namespace" sql:"not null;type:varchar(255)"`
	Repository  string     `json:"repository" sql:"not null;type:varchar(255)"`
	Description string     `json:"description" sql:"null;type:text"`
	Keys        string     `json:"keys" sql:"null;type:text"`
	Size        int64      `json:"size" sql:"default:0"`
	Locked      bool       `json:"locked" sql:"default:false"`
	CreatedAt   time.Time  `json:"created" sql:""`
	UpdatedAt   time.Time  `json:"updated" sql:""`
	DeletedAt   *time.Time `json:"deleted" sql:"index"`
}

//
func (*AppcV1) TableName() string {
	return "appc_v1"
}

//
type ACIv1 struct {
	Id        int64      `json:"id" gorm:"primary_key"`
	AppcV1    int64      `json:"appcv1" sql:"not null"`
	OS        string     `json:"os" sql:"null;type:varchar(255)"`
	Arch      string     `json:"arch" sql:"null;type:varchar(255)"`
	Name      string     `json:"name" sql:"not null;type:text"`
	OSS       string     `json:"name" sql:"null;type:text"`
	Path      string     `json:"arch" sql:"null;type:text"`
	Size      int64      `json:"size" sql:"default:0"`
	Locked    bool       `json:"locked" sql:"default:false"`
	CreatedAt time.Time  `json:"created" sql:""`
	UpdatedAt time.Time  `json:"updated" sql:""`
	DeletedAt *time.Time `json:"deleted" sql:"index"`
}

func (*ACIv1) TableName() string {
	return "aci_v1"
}
