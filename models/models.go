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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/containerops/configure"
)

var (
	db *gorm.DB
)

//
func init() {
}

//OpenDatabase is
func OpenDatabase() {
	var err error
	if db, err = gorm.Open(configure.GetString("database.driver"), configure.GetString("database.url")); err != nil {
		log.Fatal("Initlization database connection error.")
		os.Exit(1)
	} else {
		db.DB()
		db.DB().Ping()
		db.DB().SetMaxIdleConns(10)
		db.DB().SetMaxOpenConns(100)
		db.SingularTable(true)
	}
}

//Sync is
func Sync() error {
	log.Info("Sync database structs")

	db.AutoMigrate(&AppcV1{}, &ACIv1{})
	db.AutoMigrate(&AppV1{}, &ArtifactV1{})
	db.AutoMigrate(&DockerV1{}, &DockerImageV1{}, &DockerTagV1{})
	db.AutoMigrate(&DockerV2{}, &DockerImageV2{}, &DockerTagV2{})
	db.AutoMigrate(&ImageV1{}, &VirtualV1{})

	return nil
}
