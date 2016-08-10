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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/module/signature"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
)

//GetPingV2Handler is https://github.com/docker/distribution/blob/master/docs/spec/api.md#api-version-check
func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	if len(ctx.Req.Header.Get("Authorization")) == 0 {
		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Resp.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%v\"", setting.Domains))

		//TODO Bearer Token

		result, _ := json.Marshal(map[string]string{})
		return http.StatusUnauthorized, result
	}

	//TODO Decode baic authorizate data in HEADER ["Authorization"]
	//TODO Authenticate with crew project.
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetCatalogV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func HeadBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PostBlobsV2Handler is
//Initiate a resumable blob upload. If successful, an upload location will be provided to complete the upload.
//Optionally, if the digest parameter is present, the request body will be used to complete the upload in a single request.
func PostBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	//TODO: If standalone == true, Dockyard will check HEADER Authorization; if standalone == false, Dockyard will check HEADER TOEKN.
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	uuid := utils.MD5(uuid.NewV4().String())
	state := utils.MD5(fmt.Sprintf("%s/%s/%d", namespace, repository, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.Domains, namespace, repository, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", "0-0")

	result, _ := json.Marshal(map[string]string{})
	return http.StatusAccepted, result
}

//PatchBlobsV2Handler is
//Upload a chunk of data for the specified upload.
//Docker 1.9.x above version saves layer in PATCH methord
//Docker 1.9.x below version saves layer in PUT methord
func PatchBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	if upload, err := module.CheckDockerVersion19(ctx.Req.Header.Get("User-Agent")); err != nil {
		log.Errorf("Decode docker version error: %s", err.Error())

		result, _ := json.Marshal(map[string]string{})
		return http.StatusBadRequest, result
	} else if upload == true {

		//It's run above docker 1.9.0
		basePath := setting.DockerV2Storage
		uuidPath := fmt.Sprintf("%s/uuid/%s", basePath, uuid)
		uuidFile := fmt.Sprintf("%s/uuid/%s/%s", basePath, uuid, uuid)

		if !utils.IsDirExist(uuidPath) {
			os.MkdirAll(uuidPath, os.ModePerm)
		}

		if _, err := os.Stat(uuidFile); err == nil {
			os.Remove(uuidFile)
		}

		data, _ := ctx.Req.Body().Bytes()
		if err := ioutil.WriteFile(uuidFile, data, 0777); err != nil {
			log.Errorf("Save the temp file %s error: %s", uuidFile, err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}

		ctx.Resp.Header().Set("Range", fmt.Sprintf("0-%v", len(data)-1))
	}

	state := utils.MD5(fmt.Sprintf("%s/%v", fmt.Sprintf("%s/%s", namespace, repository), time.Now().UnixNano()/int64(time.Millisecond)))

	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.Domains,
		namespace,
		repository,
		uuid,
		state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutBlobsV2Handler is
//Complete the upload specified by uuid, optionally appending the body as the final chunk.
func PutBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	var size int64

	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]
	digest := ctx.Query("digest")
	tarsum := strings.Split(digest, ":")[1]

	basePath := setting.DockerV2Storage
	imagePath := fmt.Sprintf("%s/uuid/%s", basePath, tarsum)
	imageFile := fmt.Sprintf("%s/uuid/%s/%s", basePath, tarsum, tarsum)

	//Save from uuid path save too image path
	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(imageFile); err == nil {
		os.Remove(imageFile)
	}

	if upload, err := module.CheckDockerVersion19(ctx.Req.Header.Get("User-Agent")); err != nil {
		log.Errorf("Decode docker version error: %s", err.Error())

		result, _ := json.Marshal(map[string]string{})
		return http.StatusBadRequest, result
	} else if upload == true {
		//Docker 1.9.x above version saves layer in PATCH method, in PUT method move from uuid to image:sha256
		uuidFile := fmt.Sprintf("%s/uuid/%s/%s", basePath, uuid, uuid)

		var data []byte
		if _, err := os.Stat(uuidFile); err == nil {
			data, _ = ioutil.ReadFile(uuidFile)
			if err := ioutil.WriteFile(imageFile, data, 0777); err != nil {
				log.Errorf("Move the temp file to image folder %s error: %s", imageFile, err.Error())

				result, _ := json.Marshal(map[string]string{})
				return http.StatusBadRequest, result
			}
			size = int64(len(data))
			os.RemoveAll(uuidFile)
		}
	} else if upload == false {
		//Docker 1.9.x below version saves layer in PUT methord, save data to file directly.
		data, _ := ctx.Req.Body().Bytes()
		if err := ioutil.WriteFile(imageFile, data, 0777); err != nil {
			log.Errorf("Save the file %s error: %s", imageFile, err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}
		size = int64(len(data))
	}

	i := new(models.DockerImageV2)
	if err := i.Put(tarsum, imageFile, size); err != nil {
		log.Errorf("Save the iamge data %s error: %s", tarsum, err.Error())

		result, _ := json.Marshal(map[string]string{})
		return http.StatusBadRequest, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%v", fmt.Sprintf("%s/%s", namespace, repository), time.Now().UnixNano()/int64(time.Millisecond)))

	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.Domains,
		namespace,
		repository,
		uuid,
		state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutManifestsV2Handler is
func PutManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")
	agent := ctx.Req.Header.Get("User-Agent")
	tag := ctx.Params(":tag")

	if data, err := ctx.Req.Body().String(); err != nil {
		log.Errorf("Get the manifest data error: %s", err.Error())

		result, _ := json.Marshal(map[string]string{})
		return http.StatusBadRequest, result
	} else {
		_, version, _ := module.GetTarsumlist([]byte(data))
		digest, _ := signature.DigestManifest([]byte(data))

		r := new(models.DockerV2)
		if err := r.Put(namespace, repository, agent, string(version)); err != nil {
			log.Errorf("Put the manifest data error: %s", err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}

		t := new(models.DockerTagV2)
		if err := t.Put(namespace, repository, tag, "", data, version); err != nil {
			log.Errorf("Put the manifest data error: %s", err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}

		random := fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s",
			setting.Domains,
			namespace,
			repository,
			digest)

		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Resp.Header().Set("Docker-Content-Digest", digest)
		ctx.Resp.Header().Set("Location", random)

		status := []int{http.StatusBadRequest, http.StatusAccepted, http.StatusCreated}

		result, _ := json.Marshal(map[string]string{})
		return status[version], result
	}
}

func GetTagsListV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func DeleteBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func DeleteManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
