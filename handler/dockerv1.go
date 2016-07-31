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

	"gopkg.in/macaron.v1"
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

//PutRepositoryV1Handler
func PutRepositoryV1Handler(ctx *macaron.Context) (int, []byte) {
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
