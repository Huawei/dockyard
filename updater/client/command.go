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

package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "initiate default setting",
	Action: func(context *cli.Context) {
		var duc dyUpdaterConfig
		if err := duc.Init(); err == nil {
			fmt.Println("Success in initiating Dockyard Updater Client configuration.")
		} else {
			fmt.Println(err)
		}
	},
}
