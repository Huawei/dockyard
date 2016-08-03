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
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/urfave/cli"

	cutils "github.com/containerops/dockyard/cmd/client"
)

var CmdClient = cli.Command{
	Name:        "client",
	Usage:       "dockyard update service client",
	Description: "A dockyard client to pull/push/verify image/vm/app.",
	Subcommands: []cli.Command{
		initCommand,
		addCommand,
		removeCommand,
		listCommand,
		pushCommand,
		pullCommand,
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "initiate default setting",
	Action: func(context *cli.Context) error {
		var ucc cutils.UpdateClientConfig

		if err := ucc.Init(); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("Success in initiating Dockyard Updater Client configuration.")
		return nil
	},
}

var addCommand = cli.Command{
	Name:  "add",
	Usage: "add a repository url",

	Action: func(context *cli.Context) error {
		var ucc cutils.UpdateClientConfig

		repo, err := cutils.NewUCRepo(context.Args().Get(0))
		if err != nil {
			fmt.Println(err)
			return err
		}

		if err := ucc.Add(repo.String()); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Success in adding %s.\n", repo.String())
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a repository url",

	Action: func(context *cli.Context) error {
		var ucc cutils.UpdateClientConfig

		repo, err := cutils.NewUCRepo(context.Args().Get(0))
		if err != nil {
			fmt.Println(err)
			return err
		}

		if err := ucc.Remove(repo.String()); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Success in removing %s.\n", repo.String())
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "list",
	Usage: "list the saved repositories or appliances of a certain repository",

	Action: func(context *cli.Context) error {
		var ucc cutils.UpdateClientConfig

		if len(context.Args()) == 0 {
			if err := ucc.Load(); err != nil {
				fmt.Println(err)
				return err
			}

			for _, repo := range ucc.Repos {
				fmt.Println(repo)
			}
		} else if len(context.Args()) == 1 {
			uc := new(cutils.UpdateClient)
			repoURL := context.Args().Get(0)
			apps, err := uc.List(repoURL)
			if err != nil {
				fmt.Println(err)
				return err
			}

			for _, app := range apps {
				fmt.Println(app)
			}
			ucc.Add(repoURL)
		}
		return nil
	},
}

var pushCommand = cli.Command{
	Name:  "push",
	Usage: "push a file to a repository",

	Action: func(context *cli.Context) error {
		//TODO: we can have a default repo
		if len(context.Args()) != 2 {
			err := errors.New("wrong syntax: push 'filepath' 'repo url'")
			fmt.Println(err)
			return err
		}

		repo, err := cutils.NewUCRepo(context.Args().Get(1))
		if err != nil {
			fmt.Println(err)
			return err
		}

		file := context.Args().Get(0)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = repo.Put(filepath.Base(file), content)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	},
}

var pullCommand = cli.Command{
	Name:  "pull",
	Usage: "pull a file from a repository",

	Action: func(context *cli.Context) error {
		//TODO: we can have a default repo
		if len(context.Args()) != 2 {
			err := errors.New("wrong syntax: pull 'repo url' 'filename'")
			fmt.Println(err)
			return err
		}

		repoURL := context.Args().Get(0)
		fileName := context.Args().Get(1)
		uc := new(cutils.UpdateClient)

		fmt.Println("start to download file: ", fileName)
		localFile, err := uc.GetFile(repoURL, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("file downloaded to: ", localFile)

		fmt.Println("start to download public key")
		pubFile, err := uc.GetPublicKey(repoURL)
		if err != nil {
			fmt.Println("Fail to get public key: ", err)
			return err
		}
		fmt.Println("success in downloading public key to: ", pubFile)

		fmt.Println("start to download meta data and signature file")
		metaFile, err := uc.GetMeta(repoURL)
		if err != nil {
			fmt.Println("Fail to get meta data: ", err)
			return err
		}
		signFile, err := uc.GetMetaSign(repoURL)
		if err != nil {
			fmt.Println("Fail to get sign data: ", err)
			return err
		}
		fmt.Println("success in downloading meta data and signature file: %s %s", metaFile, signFile)

		//TODO
		//fmt.Println("start to verify meta data and downloaded file")
		return err
	},
}
