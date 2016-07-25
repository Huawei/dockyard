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

package handler

import (
	"encoding/json"
	"net/http"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/updater/server/utils"
)

type httpListRet struct {
	Message string
	Content interface{}
}

// List all the files in the namespace/repository
func AppListFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := utils.NewDUSProtocal("appV1")
	apps, _ := appV1.List(namespace + "/" + repository)
	ret := httpListRet{
		Message: "AppV1 List files",
		Content: apps,
	}
	result, _ := json.Marshal(ret)
	return http.StatusOK, result
}

// Get the meta data of all the namespace/repository
func AppGetMetaV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := utils.NewDUSProtocal("appV1")
	metas, _ := appV1.GetMeta(namespace + "/" + repository)
	ret := httpListRet{
		Message: "AppV1 Get Meta data",
		Content: metas,
	}
	result, _ := json.Marshal(ret)
	return http.StatusOK, result
}

// Get the content of a certain app
func AppGetFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	name := ctx.Params(":name")

	appV1, _ := utils.NewDUSProtocal("appV1")
	data, _ := appV1.Get(namespace + "/" + repository + "/" + name)
	return http.StatusOK, data
}

// Post the content of a certain app
func AppPostFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	name := ctx.Params(":name")

	data, _ := ctx.Req.Body().Bytes()
	appV1, _ := utils.NewDUSProtocal("appV1")
	appV1.Put(namespace+"/"+repository+"/"+name, data)

	ret := httpListRet{
		Message: "AppV1 Post data",
	}
	result, _ := json.Marshal(ret)
	return http.StatusOK, result
}
