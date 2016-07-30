/*
Copyright 2016 The ContainerOps Authors All rights reserved.

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

package appV1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	duc_utils "github.com/containerops/dockyard/updater/client/utils"
)

const (
	appV1Prefix  = "appV1"
	appV1Restful = "app/v1"
)

var (
	repoRegexp = regexp.MustCompile(`^(.+)://(.+)/(.+)/(.+)$`)
)

type DyUpdaterClientAppV1Repo struct {
	Site      string
	Namespace string
	Repo      string
}

func init() {
	duc_utils.RegisterRepo(appV1Prefix, &DyUpdaterClientAppV1Repo{})
}

func (ap *DyUpdaterClientAppV1Repo) Supported(url string) bool {
	return strings.HasPrefix(url, appV1Prefix+"://")
}

// New parses 'app://containerops.me/containerops/offical' and get
//	Site:       "containerops.me"
//      Namespace:  "containerops"
//      Repo:       "offical"
func (ap *DyUpdaterClientAppV1Repo) New(url string) (duc_utils.DyUpdaterClientRepo, error) {
	parts := repoRegexp.FindStringSubmatch(url)
	if len(parts) != 5 || parts[1] != appV1Prefix {
		return nil, duc_utils.ErrorsDURRepoInvalid
	}

	ap.Site = parts[2]
	ap.Namespace = parts[3]
	ap.Repo = parts[4]

	return ap, nil
}

// 'namespace/repo'
func (ap DyUpdaterClientAppV1Repo) NRString() string {
	return fmt.Sprintf("%s/%s", ap.Namespace, ap.Repo)
}

func (ap DyUpdaterClientAppV1Repo) String() string {
	return fmt.Sprintf("%s://%s/%s/%s", appV1Prefix, ap.Site, ap.Namespace, ap.Repo)
}

func (ap DyUpdaterClientAppV1Repo) generateURL() string {
	//FIXME: only support http
	return fmt.Sprintf("http://%s/%s/%s/%s", ap.Site, appV1Restful, ap.Namespace, ap.Repo)
}

func (ap DyUpdaterClientAppV1Repo) List() ([]string, error) {
	url := ap.generateURL()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	type httpRet struct {
		Message string
		Content []string
	}

	var ret httpRet
	err = json.Unmarshal(resp_body, &ret)
	if err != nil {
		return nil, err
	}

	return ret.Content, nil
}

func (ap DyUpdaterClientAppV1Repo) GetFile(name string) ([]byte, error) {
	url := fmt.Sprintf("%s/blob/%s", ap.generateURL(), name)
	return ap.getFromURL(url)
}

func (ap DyUpdaterClientAppV1Repo) GetMetaSign() ([]byte, error) {
	url := fmt.Sprintf("%s/metasign", ap.generateURL())
	return ap.getFromURL(url)
}

func (ap DyUpdaterClientAppV1Repo) GetMeta() ([]byte, error) {
	url := fmt.Sprintf("%s/meta", ap.generateURL())
	return ap.getFromURL(url)
}

func (ap DyUpdaterClientAppV1Repo) GetPublicKey() ([]byte, error) {
	url := fmt.Sprintf("%s/pubkey", ap.generateURL())
	return ap.getFromURL(url)
}

func (ap DyUpdaterClientAppV1Repo) getFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return resp_body, nil
}

func (ap DyUpdaterClientAppV1Repo) Put(name string, content []byte) error {
	url := fmt.Sprintf("%s/%s", ap.generateURL(), name)
	body := bytes.NewBuffer(content)
	resp, err := http.Post(url, "application/appv1", body)
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(resp.Body)
	return err
}
