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

package module

import (
	"encoding/json"
	"strconv"
	"strings"
)

//CheckDockerVersion19 is
func CheckDockerVersion19(headers string) (bool, error) {
	agents := map[string]string{}
	for _, v := range strings.Split(headers, " ") {
		if len(strings.Split(v, "/")) > 1 {
			agents[strings.Split(v, "/")[0]] = strings.Split(v, "/")[1]
		}
	}

	versions := strings.Split(agents["docker"], ".")
	major, _ := strconv.ParseInt(versions[0], 10, 64)
	version, _ := strconv.ParseInt(versions[1], 10, 64)

	if major > 1 {
		return true, nil
	} else if major == 1 {
		if version > 9 {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}

//GetTarsumlist is
func GetTarsumlist(data []byte) ([]string, int64, error) {
	var tarsumlist []string
	var layers = []string{"", "fsLayers", "layers"}
	var tarsums = []string{"", "blobSum", "digest"}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return []string{}, 0, err
	}

	schemaVersion := int64(manifest["schemaVersion"].(float64))

	if schemaVersion == 2 {
		confblobsum := manifest["config"].(map[string]interface{})["digest"].(string)
		tarsum := strings.Split(confblobsum, ":")[1]
		tarsumlist = append(tarsumlist, tarsum)
	}

	section := layers[schemaVersion]
	item := tarsums[schemaVersion]
	for i := len(manifest[section].([]interface{})) - 1; i >= 0; i-- {
		blobsum := manifest[section].([]interface{})[i].(map[string]interface{})[item].(string)
		tarsum := strings.Split(blobsum, ":")[1]
		tarsumlist = append(tarsumlist, tarsum)
	}

	return tarsumlist, schemaVersion, nil
}
