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

func HeadBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PatchBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutManifestsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetTagsListV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func GetManifestsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func DeleteBlobsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func DeleteManifestsV2LibraryHandler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
