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
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
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

//GetCatalogV2Handler is
func GetCatalogV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//HeadBlobsV2Handler is
func HeadBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.DockerImageV2)
	if err := i.Get(tarsum); err != nil && err == gorm.ErrRecordNotFound {
		log.Info("Not found blob: %s", tarsum)

		result, _ := module.EncodingError(module.BLOB_UNKNOWN, digest)
		return http.StatusNotFound, result
	} else if err != nil && err != gorm.ErrRecordNotFound {
		log.Info("Failed to get blob %s: %s", tarsum, err.Error())

		result, _ := module.EncodingError(module.UNKNOWN, err.Error())
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(i.Size))

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

	r := new(models.DockerV2)

	if err := r.Put(namespace, repository); err != nil {
		log.Errorf("Put or search repository error: %s", err.Error())

		result, _ := module.EncodingError(module.UNKNOWN, map[string]string{"namespace": namespace, "repository": repository})
		return http.StatusBadRequest, result
	}

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

		result, _ := module.EncodingError(module.BLOB_UPLOAD_UNKNOWN, map[string]string{"namespace": namespace, "repository": repository})
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

			result, _ := module.EncodingError(module.BLOB_UPLOAD_UNKNOWN, map[string]string{"namespace": namespace, "repository": repository})
			return http.StatusBadRequest, result
		}

		ctx.Resp.Header().Set("Range", fmt.Sprintf("0-%v", len(data)-1))
	}

	state := utils.MD5(fmt.Sprintf("%s/%v", fmt.Sprintf("%s/%s", namespace, repository), time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.Domains, namespace, repository, uuid, state)

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
	imagePath := fmt.Sprintf("%s/image/%s", basePath, tarsum)
	imageFile := fmt.Sprintf("%s/image/%s/%s", basePath, tarsum, tarsum)

	//Save from uuid path save too image path
	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(imageFile); err == nil {
		os.Remove(imageFile)
	}

	if upload, err := module.CheckDockerVersion19(ctx.Req.Header.Get("User-Agent")); err != nil {
		log.Errorf("Decode docker version error: %s", err.Error())

		result, _ := module.EncodingError(module.BLOB_UPLOAD_INVALID, map[string]string{"namespace": namespace, "repository": repository})
		return http.StatusBadRequest, result
	} else if upload == true {
		//Docker 1.9.x above version saves layer in PATCH method, in PUT method move from uuid to image:sha256
		uuidPath := fmt.Sprintf("%s/uuid/%s", basePath, uuid)
		uuidFile := fmt.Sprintf("%s/uuid/%s/%s", basePath, uuid, uuid)

		var data []byte
		if _, err := os.Stat(uuidFile); err == nil {
			data, _ = ioutil.ReadFile(uuidFile)
			if err := ioutil.WriteFile(imageFile, data, 0777); err != nil {
				log.Errorf("Move the temp file to image folder %s error: %s", imageFile, err.Error())

				result, _ := module.EncodingError(module.BLOB_UPLOAD_INVALID, map[string]string{"namespace": namespace, "repository": repository})
				return http.StatusBadRequest, result
			}

			size = int64(len(data))

			os.RemoveAll(uuidFile)
			os.RemoveAll(uuidPath)
		}
	} else if upload == false {
		//Docker 1.9.x below version saves layer in PUT methord, save data to file directly.
		data, _ := ctx.Req.Body().Bytes()
		if err := ioutil.WriteFile(imageFile, data, 0777); err != nil {
			log.Errorf("Save the file %s error: %s", imageFile, err.Error())

			result, _ := module.EncodingError(module.BLOB_UPLOAD_INVALID, map[string]string{"namespace": namespace, "repository": repository})
			return http.StatusBadRequest, result
		}

		size = int64(len(data))
	}

	i := new(models.DockerImageV2)
	if err := i.Put(tarsum, imageFile, size); err != nil {
		log.Errorf("Save the iamge data %s error: %s", tarsum, err.Error())

		result, _ := module.EncodingError(module.BLOB_UPLOAD_INVALID, map[string]string{"namespace": namespace, "repository": repository})
		return http.StatusBadRequest, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%v", fmt.Sprintf("%s/%s", namespace, repository), time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.Domains, namespace, repository, uuid, state)

	ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetBlobsV2Handler is
func GetBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	var file []byte
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.DockerImageV2)
	if err := i.Get(tarsum); err != nil && err == gorm.ErrRecordNotFound {
		log.Info("Not found blob: %s", tarsum)

		result, _ := module.EncodingError(module.BLOB_UNKNOWN, digest)
		return http.StatusNotFound, result
	} else if err != nil && err != gorm.ErrRecordNotFound {
		log.Info("Failed to get blob %s: %s", tarsum, err.Error())

		result, _ := module.EncodingError(module.UNKNOWN, err.Error())
		return http.StatusBadRequest, result
	}

	if data, err := ioutil.ReadFile(i.Path); err != nil {
		log.Info("Failed to get blob %s: %s", tarsum, err.Error())

		result, _ := module.EncodingError(module.UNKNOWN, err.Error())
		return http.StatusBadRequest, result
	} else {
		file = data
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
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
		_, imageID, version, _ := module.GetTarsumlist([]byte(data))
		digest, _ := signature.DigestManifest([]byte(data))

		r := new(models.DockerV2)
		if err := r.PutAgent(namespace, repository, agent, strconv.FormatInt(version, 10)); err != nil {
			log.Errorf("Put the manifest data error: %s", err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}

		t := new(models.DockerTagV2)
		if err := t.Put(namespace, repository, tag, imageID, data, strconv.FormatInt(version, 10)); err != nil {
			log.Errorf("Put the manifest data error: %s", err.Error())

			result, _ := json.Marshal(map[string]string{})
			return http.StatusBadRequest, result
		}

		random := fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s",
			setting.Domains, namespace, repository, digest)

		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctx.Resp.Header().Set("Docker-Content-Digest", digest)
		ctx.Resp.Header().Set("Location", random)

		status := []int{http.StatusBadRequest, http.StatusAccepted, http.StatusCreated}

		result, _ := json.Marshal(map[string]string{})
		return status[version], result
	}
}

//GetTagsListV2Handler is
func GetTagsListV2Handler(ctx *macaron.Context) (int, []byte) {
	var err error
	repository := ctx.Params(":repository")
	namespace := ctx.Params(":namespace")

	data := map[string]interface{}{}
	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	r := new(models.DockerV2)

	if data["tags"], err = r.GetTags(namespace, repository); err != nil && err == gorm.ErrRecordNotFound {
		log.Info("Not found tags: %s/%s", namespace, repository)

		result, _ := module.EncodingError(module.BLOB_UNKNOWN, fmt.Sprintf("%s/%s", namespace, repository))
		return http.StatusNotFound, result
	} else if err != nil && err != gorm.ErrRecordNotFound {
		log.Info("Failed to get tags %s/%s: %s", namespace, repository, err.Error())

		result, _ := module.EncodingError(module.UNKNOWN, err.Error())
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(data)
	return http.StatusOK, result
}

//GetManifestsV2Handler is
func GetManifestsV2Handler(ctx *macaron.Context) (int, []byte) {

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//DeleteBlobsV2Handler is
func DeleteBlobsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//DeleteManifestsV2Handler is
func DeleteManifestsV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
