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

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/setting"
)

type httpListRet struct {
	Message string
	Content interface{}
}

//TODO: better http return result
func httpRet(head string, content interface{}, err error) (int, []byte) {
	var ret httpListRet
	var code int

	if err != nil {
		ret.Message = head + " fail"
		ret.Content = err.Error()
		code = http.StatusBadRequest
	} else {
		ret.Message = head
		ret.Content = content
		code = http.StatusOK
	}

	result, _ := json.Marshal(ret)
	return code, result
}

// List all the files in the namespace/repository
func AppListFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	apps, err := appV1.List(namespace + "/" + repository)

	return httpRet("AppV1 List files", apps, err)
}

// Get the public key of the namespace/repository
func AppGetPublicKeyV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetPublicKey(namespace + "/" + repository)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Public Key", nil, err)
	}
}

// Get the meta data of all the namespace/repository
func AppGetMetaV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetMeta(namespace + "/" + repository)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Meta", nil, err)
	}
}

// Get the meta signature data of all the namespace/repository
func AppGetMetaSignV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetMetaSign(namespace + "/" + repository)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Meta Sign", data, err)
	}
}

// Get the content of a certain app
func AppGetFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	name := ctx.Params(":name")

	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.Get(namespace+"/"+repository, name)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get File", nil, err)
	}
}

// Post the content of a certain app
func AppPostFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	name := ctx.Params(":name")

	data, _ := ctx.Req.Body().Bytes()
	appV1, _ := module.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	err := appV1.Put(namespace+"/"+repository, name, data)

	return httpRet("AppV1 Post data", nil, err)
}
