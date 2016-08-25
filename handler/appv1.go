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

package handler

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/updateservice/us"
	"github.com/containerops/dockyard/utils"
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

//Example: curl https://containerops.me/app/v1/search?namespace=genedna&repository=tidb
func AppGlobalSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func AppDiscoveryV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//Example: curl https://containerops.me/app/v1/genedna/tidb/search?version=beta
func AppScopedSearchV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// AppGetListAppV1Handler lists all the files in the namespace/repository
func AppGetListAppV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	apps, err := appV1.List(namespace + "/" + repository)

	return httpRet("AppV1 List files", apps, err)
}

// AppGetPublicKeyV1Handler gets the public key of the namespace/repository
func AppGetPublicKeyV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")

	appV1, _ := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetPublicKey(namespace)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Public Key", nil, err)
	}
}

// AppGetMetaV1Handler gets the meta data of the whole namespace/repository
func AppGetMetaV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetMeta(namespace + "/" + repository)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Meta", nil, err)
	}
}

// AppGetMetaSignV1Handler gets the meta signature data
func AppGetMetaSignV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	appV1, _ := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.GetMetaSign(namespace + "/" + repository)
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get Meta Sign", data, err)
	}
}

// AppGetFileV1Handler gets the data of a certain app
func AppGetFileV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	a := models.ArtifactV1{
		OS:   ctx.Params(":os"),
		Arch: ctx.Params(":arch"),
		App:  ctx.Params(":app"),
		Tag:  ctx.Params(":tag"),
	}

	appV1, _ := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	data, err := appV1.Get(namespace+"/"+repository, a.GetName())
	if err == nil {
		return http.StatusOK, data
	} else {
		return httpRet("AppV1 Get File", nil, err)
	}
}

//
func AppGetManifestsV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//
func AppPostFileV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// AppPutFileV1Handler creates or updates a certain app
func AppPutFileV1Handler(ctx *macaron.Context) (int, []byte) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("[%s] Req.Body.Bytes error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Req.Body.Bytes Error"})
		return http.StatusBadRequest, result
	}

	// Query or Create the repository.
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	r, err := models.NewAppV1(namespace, repository)
	if err != nil {
		log.Errorf("[%s] query/create repository error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Query/Create Repository Error"})
		return http.StatusBadRequest, result
	}

	reqMethod := ctx.Req.Header.Get("Dockyard-Encrypt-Method")
	encryptMethod := utils.NewEncryptMethod(reqMethod)
	if encryptMethod == utils.EncryptNotSupported {
		log.Errorf("[%s] encrypt method %s is invalid", ctx.Req.RequestURI, reqMethod)

		result, _ := json.Marshal(map[string]string{"Error": "Invalid Encrypt Method"})
		return http.StatusBadRequest, result
	}

	a := models.ArtifactV1{
		OS:            ctx.Params(":os"),
		Arch:          ctx.Params(":arch"),
		App:           ctx.Params(":app"),
		Tag:           ctx.Params(":tag"),
		EncryptMethod: string(encryptMethod),
		Size:          int64(len(data)),
	}

	// Add to update service
	appV1, err := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	if err != nil {
		log.Errorf("[%s] create update service: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Create Update Service Error"})
		return http.StatusBadRequest, result
	}

	tmpPath, err := appV1.Put(namespace+"/"+repository, a.GetName(), data, encryptMethod)
	if err != nil {
		log.Errorf("[%s] put to update service error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Update Service Error"})
		return http.StatusBadRequest, result
	}

	// Although we record the local storage path (or object storage key), we do load it by UpdateService.
	a.Path = tmpPath
	err = r.Put(a)
	if err != nil {
		log.Errorf("[%s] put artifact error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "PUT Artifact Error"})
		return http.StatusBadRequest, result
	}

	return httpRet("AppV1 Put data", nil, err)
}

//
func AppPutManifestV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//
func AppPatchFileV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

// AppDeleteFileV1Handler remove a file from a repo
func AppDeleteFileV1Handler(ctx *macaron.Context) (int, []byte) {
	// setup the repository.
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	r, err := models.NewAppV1(namespace, repository)
	if err != nil {
		log.Errorf("[%s] setup repository error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Setup Repository Error"})
		return http.StatusBadRequest, result
	}

	a := models.ArtifactV1{
		OS:   ctx.Params(":os"),
		Arch: ctx.Params(":arch"),
		App:  ctx.Params(":app"),
		Tag:  ctx.Params(":tag"),
	}

	// Remove from update service
	appV1, err := us.NewUpdateService("appV1", setting.Storage, setting.KeyManager)
	if err != nil {
		log.Errorf("[%s] create update service: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Create Update Service Error"})
		return http.StatusBadRequest, result
	}

	err = appV1.Delete(namespace+"/"+repository, a.GetName())
	if err != nil {
		log.Errorf("[%s] delete from update service error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Delete Update Service Error"})
		return http.StatusBadRequest, result
	}

	err = r.Delete(a)
	if err != nil {
		log.Errorf("[%s] delete artifact error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Delete Artifact Error"})
		return http.StatusBadRequest, result
	}

	return httpRet("AppV1 Delete data", nil, err)
}

// AppRegistScanHooksV1Handler adds a scan plugin to a user repo
// TODO: to make it easier as a start, we assume each repo could only have one scan plugin
func AppRegistScanHooksV1Handler(ctx *macaron.Context) (int, []byte) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("[%s] Req.Body.Bytes error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Req.Body.Bytes Error"})
		return http.StatusBadRequest, result
	}

	type scanPlugin struct {
		Name string
	}
	var n scanPlugin
	err = json.Unmarshal(data, &n)
	if err != nil {
		log.Errorf("[%s] Invalid body data: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Parse Req.Body.Bytes Error"})
		return http.StatusBadRequest, result
	}

	var reg models.ScanHookRegist
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	err = reg.Regist("appv1", namespace, repository, n.Name)
	if err != nil {
		log.Errorf("[%s] scan hook regist error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Scan Hook Regist Error"})
		return http.StatusBadRequest, result
	}

	return httpRet("AppV1 Scan Hook Regist", nil, err)
}

// AppCallbackScanHooksV1Handler gets callback from container and save the scan result.
func AppCallbackScanHooksV1Handler(ctx *macaron.Context) (int, []byte) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Errorf("[%s] Req.Body.Bytes error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Req.Body.Bytes Error"})
		return http.StatusBadRequest, result
	}

	var t models.ScanHookTask
	callbackID := ctx.Params(":callbackID")
	err = t.UpdateResult(callbackID, data)
	if err != nil {
		log.Errorf("[%s] scan hook callback error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Scan Hook Callback Error"})
		return http.StatusBadRequest, result
	}

	return httpRet("AppV1 Scan Hook Callback", nil, err)
}

// AppActiveScanHooksTaskV1Handler actives a scan task
func AppActiveScanHooksTaskV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	var r models.ScanHookRegist
	rID, err := r.FindID("appv1", namespace, repository)
	if err != nil {
		log.Errorf("[%s] scan hook callback error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Donnot have registed scan plugin"})
		return http.StatusBadRequest, result
	}

	a := models.ArtifactV1{
		OS:   ctx.Params(":os"),
		Arch: ctx.Params(":arch"),
		App:  ctx.Params(":app"),
		Tag:  ctx.Params(":tag"),
	}
	a, err = a.Get()
	if err != nil {
		log.Errorf("[%s] scan hook callback error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Cannot find artifactv1"})
		return http.StatusBadRequest, result
	}

	// create a task
	var t models.ScanHookTask
	tID, err := t.Put(rID, a.Path)
	if err != nil {
		log.Errorf("[%s] scan hook callback error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Fail to create a scan task"})
		return http.StatusBadRequest, result
	}

	idBytes, err := utils.TokenMarshal(tID, setting.ScanKey)

	val := struct {
		TaskID string
	}{TaskID: string(idBytes)}

	return httpRet("AppV1 Active Scan Hook Task", val, nil)
}
