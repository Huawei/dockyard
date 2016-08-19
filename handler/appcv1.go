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
	"html/template"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/setting"
)

//AppcDiscoveryV1Handler is
func AppcDiscoveryV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	discovery := ctx.Query("ac-discovery")

	if len(discovery) > 0 && discovery == "1" {
		if t, err := template.ParseGlob("views/aci/discovery.html"); err != nil {
			log.Errorf("[%s] get gpg file template status: %s", ctx.Req.RequestURI, err.Error())

			result, _ := json.Marshal(map[string]string{"Error": "Get GPG File Template Status Error"})
			return http.StatusBadRequest, result
		} else {
			t.Execute(ctx.Resp, map[string]string{
				"Domains":    setting.Domains,
				"Namespace":  namespace,
				"Repository": repository,
			})
		}
	}

	return http.StatusOK, []byte("")
}

//AppcGetACIV1Handler is
func AppcGetACIV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

type AppcPUTDetails struct {
	ACIPushVersion string `json:"aci_push_version"`
	Multipart      bool   `json:"multipart"`
	ManifestURL    string `json:"upload_manifest_url"`
	SignatureURL   string `json:"upload_signature_url"`
	ACIURL         string `json:"upload_aci_url"`
	CompletedURL   string `json:"completed_url"`
}

//AppcPostACIV1Handler is
func AppcPostACIV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	aci := ctx.Params(":aci")

	version := strings.Split(aci, "-")[1]

	prefix := fmt.Sprintf("https://%s/appc/%s/%s/push", setting.Domains, namespace, repository)

	appc := AppcPUTDetails{
		ACIPushVersion: "0.0.1",
		Multipart:      false,
		ManifestURL:    fmt.Sprintf("%s/%s/manifest", prefix, version),
		SignatureURL:   fmt.Sprintf("%s/%s/asc/%s.asc", prefix, version, aci),
		ACIURL:         fmt.Sprintf("%s/%s/aci/%s", prefix, version, aci),
		CompletedURL:   fmt.Sprintf("%s/%s/complete", prefix, version),
	}

	result, _ := json.Marshal(appc)
	return http.StatusOK, result
}

func AppcPutManifestV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func AppcPutASCV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func AppcPutACIV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func AppcPostCompleteV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
