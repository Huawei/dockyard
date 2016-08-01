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
	"os"

	"github.com/urfave/cli"

	_ "github.com/containerops/dockyard/cmd/client/module/repo/appv1"
)

func main() {
	app := cli.NewApp()

	app.Name = "duc"
	app.Usage = "Dockyard Update Client"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		initCommand,
		addCommand,
		removeCommand,
		listCommand,
		pushCommand,
		pullCommand,
	}

	app.Run(os.Args)
}
