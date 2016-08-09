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
package unittest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/containerops/dockyard/utils"
)

func TestUtilsContainer(t *testing.T) {
	imageName := "ospaf/centos-ntp"
	containerName := "testcontainer"
	cached, err := utils.IsImageCached(imageName)
	if err == utils.ErrorsNoDockerClient {
		fmt.Println("Please start a docker daemon to continue the test")
		return
	}

	assert.Nil(t, err, "Fail to load Image")

	if !cached {
		err := utils.PullImage(imageName)
		assert.Nil(t, err, "Fail to pull image")
	}

	utils.StartContainer(imageName, containerName)
	//TODO: stop, remove the container process
}
