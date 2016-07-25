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

package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/containerops/dockyard/models"
)

var CmdDatabase = cli.Command{
	Name:        "database",
	Usage:       "database utils for backend database",
	Description: "Dockyard run base SQL database like MySQL, database command provide some utils of migrate, backup, config and so on.",
	Action:      runDatabase,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "action",
			Usage: "Action，[sync/backup/restore]",
		},
	},
}

func runDatabase(c *cli.Context) error {
	if len(c.String("action")) > 0 {
		action := c.String("action")

		switch action {
		case "sync":
			if err := models.Sync(); err != nil {
				fmt.Println("Init database struct error, ", err.Error())
				return err
			}
			break
		default:
			break
		}
	}

	return nil
}
