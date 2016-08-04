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

//Docker
type DockerV2 struct {
	Id            int64      `json:"id" gorm:"primary_key"`
	Namespace     string     `json:"namespace" sql:"not null;type:varchar(255)"  gorm:"unique_index:v2_repository"`
	Repository    string     `json:"repository" sql:"not null;type:varchar(255)"  gorm:"unique_index:v2_repository"`
	SchemaVersion string     `json:"schema_version" sql:"not null;type:varchar(255)"`
	Manifests     string     `json:"manifests" sql:"null;type:text"`
	Agent         string     `json:"agent" sql:"null;type:text"`
	Description   string     `json:"description" sql:"null;type:text"`
	Size          int64      `json:"size" sql:"default:0"`
	Locked        bool       `json:"locked" sql:"default:false"`
	CreatedAt     time.Time  `json:"create_at" sql:""`
	UpdatedAt     time.Time  `json:"update_at" sql:""`
	DeletedAt     *time.Time `json:"delete_at" sql:"index"`
}

//
func (*DockerV2) TableName() string {
	return "docker_V2"
}

//
type DockerImageV2 struct {
	Id              int64      `json:"id" gorm:"primary_key"`
	ImageId         string     `json:"image_id" sql:"unique;type:varchar(255)"`
	BlobSum         string     `json:"blob_sum" sql:"null;unique;type:varchar(255)"`
	V1Compatibility string     `json:"v1_compatibility" sql:"null;type:text"`
	Path            string     `json:"path" sql:"null;type:text"`
	OSS             string     `json:"oss" sql:"null;type:text"`
	Size            int64      `json:"size" sql:"default:0"`
	Locked          bool       `json:"locked" sql:"default:false"`
	CreatedAt       time.Time  `json:"create_at" sql:""`
	UpdatedAt       time.Time  `json:"update_at" sql:""`
	DeletedAt       *time.Time `json:"delete_at" sql:"index"`
}

//
func (*DockerImageV2) TableName() string {
	return "docker_image_v2"
}

//
type DockerTagV2 struct {
	Id        int64      `json:"id" gorm:"primary_key"`
	DockerV2  int64      `json:"docker_v2" sql:"not null"`
	Tag       string     `json:"tag" sql:"not null;type:varchar(255)"`
	ImageId   string     `json:"image_id" sql:"not null;type:varchar(255)"`
	Manifest  string     `json:"manifest" sql:"null;type:text"`
	Schema    int64      `json:"schema" sql:""`
	CreatedAt time.Time  `json:"create_at" sql:""`
	UpdatedAt time.Time  `json:"update_at" sql:""`
	DeletedAt *time.Time `json:"delete_at" sql:"index"`
}

//
func (*DockerTagV2) TableName() string {
	return "docker_tag_V2"
}
