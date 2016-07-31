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
	"net/http"

	"github.com/ngaut/log"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
)

//GetPingV1Handler returns http.StatusOK(200) when Dockyard provide the Docker Registry V1 support.
//TODO: Add a config option for provide Docker Registry V1.
func GetPingV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetUsersV1Handler is Docker client login handler functoin, should be integration with [Crew](https://gitub.com/containerops/crew) project.
//TODO: Integration with Crew project.
func GetUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PostUsersV1Handler In Docker Registry V1, the Docker client will POST /v1/users to create an user.
//If the Dockyard allow create user in the CLI, should be integration with [Crew](https://github.com/containerops/crew).
//If don't, Dockyard returns http.StatusUnauthorized(401) for forbidden.
//TODO: Add a config option for allow/forbidden create user in the CLI, and integrated with [Crew](https://github.com/containerops/crew).
func PostUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusUnauthorized, result
}

//PutTagV1Handler
func PutTagV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutRepositoryImagesV1Handler
func PutRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetRepositoryImagesV1Handler
func GetRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetTagV1Handler
func GetTagV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutRepositoryV1Handler will create or update the repository, it's first step of Docker push.
//TODO: @1 When someone create or update the repository, it will be locked to forbidden others action include pull action.
//TODO: @2 Add a config option for allow/forbidden Docker client pull action when a repository is locked.
//TODO: @3 Intergated with [Crew](https://github.com/containerops/crew).
//TODO: @4 Token will be store in Redis, and link the push action with username@repository.
func PutRepositoryV1Handler(ctx *macaron.Context) (int, []byte) {
	var username, body string
	//var passwd string
	var err error

	if username, _, err = utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization")); err != nil {
		log.Errorf("[%s] decode Authorization error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Decode Authorization Error"})
		return http.StatusUnauthorized, result
	}

	//When integrated with crew, like this:
	//@1: username, passwd, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	//@2: username, passwd authorizated in Crew.

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	//When integrated the Crew, should be check the privilage.
	if username != namespace {

	}

	//In Docker Registry V1, the repository json data in the body of `PUT /v1/:namespace/:repository`
	if body, err = ctx.Req.Body().String(); err != nil {
		log.Errorf("[%s] get repository json from http body error: %s", ctx.Req.RequestURI, err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Get Repository JSON Error"})
		return http.StatusBadRequest, result
	}

	//Create or update the repository.
	r := new(models.DockerV1)
	if e := r.Put(namespace, repository, body, ctx.Req.Header.Get("User-Agent")); e != nil {
		log.Errorf("[%s] put repository error: %s", ctx.Req.RequestURI, e.Error())

		result, _ := json.Marshal(map[string]string{"Error": "PUT Repository Error"})
		return http.StatusBadRequest, result
	}

	//If the Docker client use "X-Docker-Token", will return a randon token value.
	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
			utils.MD5(username), namespace, repository, "write")

		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetImageAncestryV1Handler
func GetImageAncestryV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetImageJSONV1Handler
func GetImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//GetImageLayerV1Handler
func GetImageLayerV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutImageJSONV1Handler
func PutImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutImageLayerV1Handler
func PutImageLayerV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

//PutImageChecksumV1Handler
func PutImageChecksumV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
