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

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"
)

//IndexV1Handler is
func IndexV1Handler(ctx *macaron.Context) (int, []byte) {
	discovery := ctx.Query("ac-discovery")

	if len(discovery) > 0 && discovery == "1" {
		if _, err := os.Stat("external/signs/pubkeys.gpg"); err != nil {
			log.Errorf("[%s] get gpg file status: %s", ctx.Req.RequestURI, err.Error())

			result, _ := json.Marshal(map[string]string{"Error": "Get GPG File Status Error"})
			return http.StatusBadRequest, result
		}

		if file, err := ioutil.ReadFile("external/signs/pubkeys.gpg"); err != nil {
			log.Errorf("[%s] get gpg file data: %s", ctx.Req.RequestURI, err.Error())

			result, _ := json.Marshal(map[string]string{"Error": "Get GPG File Data Error"})
			return http.StatusBadRequest, result
		} else {
			ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
			ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

			return http.StatusOK, file
		}

	}

	result, _ := json.Marshal(map[string]string{"message": "Dockyard Backend REST API Service"})
	return http.StatusOK, result
}
