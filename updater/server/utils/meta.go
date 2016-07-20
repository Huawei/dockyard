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

package utils

import (
	"fmt"
)

type Meta struct {
	Name string
}

func GenerateMeta(file string, content []byte) (meta Meta) {
	meta.Name = file
	fmt.Println("generate meta")
	return
}

func (a Meta) Compare(b Meta) int {
	if a.Name == b.Name {
		return 0
	}

	if a.Name > b.Name {
		return 1
	}

	return -1
}
