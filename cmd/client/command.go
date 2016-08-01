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
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/containerops/dockyard/cmd/client/module"
	"github.com/containerops/dockyard/utils"
)

var initCommand = cli.Command{
	Name:  "init",
	Usage: "initiate default setting",
	Action: func(context *cli.Context) error {
		var ucc UpdateClientConfig

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
		var ucc UpdateClientConfig

		repo, err := module.NewUCRepo(context.Args().Get(0))
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
		var ucc UpdateClientConfig

		repo, err := module.NewUCRepo(context.Args().Get(0))
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
		var ucc UpdateClientConfig

		if len(context.Args()) == 0 {
			if err := ucc.Load(); err != nil {
				fmt.Println(err)
				return err
			}

			for _, repo := range ucc.Repos {
				fmt.Println(repo)
			}
		} else if len(context.Args()) == 1 {
			repo, err := module.NewUCRepo(context.Args().Get(0))
			if err != nil {
				fmt.Println(err)
				return err
			}

			apps, err := repo.List()
			if err != nil {
				fmt.Println(err)
				return err
			}

			for _, app := range apps {
				fmt.Println(app)
			}
			ucc.Add(repo.String())
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

		repo, err := module.NewUCRepo(context.Args().Get(1))
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
			err := errors.New("wrong syntax: pull 'filename' 'repo url'")
			fmt.Println(err)
			return err
		}

		repo, err := module.NewUCRepo(context.Args().Get(1))
		if err != nil {
			fmt.Println(err)
			return err
		}

		var ucc UpdateClientConfig
		ucc.Init()

		file := context.Args().Get(0)
		fileBytes, err := repo.GetFile(file)
		if err != nil {
			fmt.Println(err)
			return err
		}

		localFile := filepath.Join(ucc.CacheDir, repo.NRString(), file)
		if !utils.IsDirExist(filepath.Dir(localFile)) {
			os.MkdirAll(filepath.Dir(localFile), 0755)
		}
		err = ioutil.WriteFile(localFile, fileBytes, 0644)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("file downloaded to: ", localFile)

		fmt.Println("start to download public key")
		pubBytes, err := repo.GetPublicKey()
		if err != nil {
			fmt.Println("Fail to get public key: ", err)
			return err
		}
		fmt.Println("success in downloading public key")

		fmt.Println("start to download meta data and signature file")
		metaBytes, err := repo.GetMeta()
		if err != nil {
			fmt.Println("Fail to get meta data: ", err)
			return err
		}
		signBytes, err := repo.GetMetaSign()
		if err != nil {
			fmt.Println("Fail to get sign data: ", err)
			return err
		}
		fmt.Println("success in downloading meta data and signature file")

		fmt.Println("start to verify meta data and downloaded file")
		err = utils.SHA256Verify(pubBytes, metaBytes, signBytes)
		if err != nil {
			fmt.Println("Fail to verify meta by public key")
			return err
		}
		fmt.Println("success in verifying meta data and signature file")

		fmt.Println("start to compare the hash value")
		var metas []utils.Meta
		fileHash := fmt.Sprintf("%x", sha1.Sum(fileBytes))
		json.Unmarshal(metaBytes, &metas)
		for _, meta := range metas {
			if meta.Name != file {
				continue
			}

			if meta.Hash == fileHash {
				fmt.Println("Congratulations! The file is valid!")
				return nil
			}

			err := errors.New("the file is invalid, maybe security issue")
			fmt.Println(err)
			return err
		}

		err = errors.New("something wrong with the server, cannot find the file in the meta data")
		fmt.Println(err)

		return err
	},
}
