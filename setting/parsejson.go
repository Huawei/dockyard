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

package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type NotificationsCtx struct {
	Name      string         `json:"name,omitempty"`
	Endpoints []EndpointDesc `json:"endpoints,omitempty"`
}

type EndpointDesc struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Headers   http.Header   `json:"headers"`
	Timeout   time.Duration `json:"timeout"`
	Threshold int           `json:"threshold"`
	Backoff   time.Duration `json:"backoff"`
	EventDB   string        `json:"eventdb"`
	Disabled  bool          `json:"disabled"`
}

type AuthorDesc map[string]interface{}

type AuthorsCtx map[string]AuthorDesc

type Desc struct {
	Notifications NotificationsCtx `json:"notifications,omitempty"`
	Authors       AuthorsCtx       `json:"auth,omitempty"`
}

var JSONConfCtx Desc

func GetConfFromJSON(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	if err := json.Unmarshal(buf, &JSONConfCtx); err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	return nil
}

func (auth AuthorsCtx) Name() (name string) {
	name = ""
	for key, _ := range auth {
		name = key
		break
	}
	return
}
