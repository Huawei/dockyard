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

package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/containerops/dockyard/cmd"
	// Load the local key manager module
	_ "github.com/containerops/dockyard/module/km/local"
	// Load the local storage module
	_ "github.com/containerops/dockyard/module/storage/local"
	// Load the local update service module
	_ "github.com/containerops/dockyard/module/us/appv1"
	"github.com/containerops/dockyard/setting"
)

func init() {
	//
}

func main() {
	if setting.RunMode == "prod" {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stderr)
	}

	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
		cmd.CmdDatabase,
		cmd.CmdOSS,
		cmd.CmdMonitor,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
