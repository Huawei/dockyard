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

	"github.com/urfave/cli"

	"github.com/containerops/dockyard/updater/client/utils"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "initiate default setting",
	Action: func(context *cli.Context) error {
		var duc utils.DyUpdaterClientConfig

		if err := duc.Init(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Success in initiating Dockyard Updater Client configuration.")
		}

		return nil
	},
}

var addCommand = cli.Command{
	Name:  "add",
	Usage: "add a repository url",

	Action: func(context *cli.Context) error {
		var duc utils.DyUpdaterClientConfig

		if repo, err := utils.NewDUCRepo(context.Args().Get(0)); err != nil {
			fmt.Println(err)
		} else if err := duc.Add(repo.String()); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Success in adding %s.\n", repo.String())
		}

		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a repository url",

	Action: func(context *cli.Context) error {
		var duc utils.DyUpdaterClientConfig

		if repo, err := utils.NewDUCRepo(context.Args().Get(0)); err != nil {
			fmt.Println(err)
		} else if err := duc.Remove(repo.String()); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Success in removing %s.\n", repo.String())
		}

		return nil
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list the saved repositories or appliances of a certain repository",

	Action: func(context *cli.Context) error {
		var duc utils.DyUpdaterClientConfig

		if len(context.Args()) == 0 {
			if err := duc.Load(); err != nil {
				fmt.Println(err)
			} else {
				for _, repo := range duc.Repos {
					fmt.Println(repo)
				}
			}
		} else if len(context.Args()) == 1 {
			if repo, err := utils.NewDUCRepo(context.Args().Get(0)); err != nil {
				fmt.Println(err)
			} else {
				apps, _ := repo.List()
				for _, app := range apps {
					fmt.Println(app)
				}
				duc.Add(repo.String())
			}
		}

		return nil
	},
}
