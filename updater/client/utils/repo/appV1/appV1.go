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
	"fmt"
	"regexp"
	"strings"

	duc_utils "github.com/containerops/dockyard/updater/client/utils"
)

const (
	appV1Prefix = "appV1"
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

func (ap *DyUpdaterClientAppV1Repo) String() string {
	return fmt.Sprintf("%s://%s/%s/%s", appV1Prefix, ap.Site, ap.Namespace, ap.Repo)
}

func (ap *DyUpdaterClientAppV1Repo) Get() error {
	return nil
}

func (ap *DyUpdaterClientAppV1Repo) List() ([]string, error) {
	return nil, nil
}
