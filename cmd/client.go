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
	"github.com/containerops/dockyard/updateservice/client"
	"github.com/containerops/dockyard/utils"
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
		deleteCommand,
		decryptCommand,
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

		url := context.Args().Get(0)
		if err := ucc.Add(url); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Success in adding %s.\n", url)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a repository url",

	Action: func(context *cli.Context) error {
		var ucc cutils.UpdateClientConfig

		url := context.Args().Get(0)
		if err := ucc.Remove(url); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Success in removing %s.\n", url)
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
		//TODO: we can have a default server
		var encrypt utils.EncryptMethod
		argsLen := len(context.Args())
		if argsLen == 4 {
			method := context.Args().Get(3)
			encrypt = utils.NewEncryptMethod(method)
			if encrypt == utils.EncryptNotSupported {
				err := fmt.Errorf("Encrypt method %s is not supported, should neither be 'none' or 'gpg'", method)
				fmt.Println(err)
				return err
			}
		} else {
			encrypt = utils.EncryptNone
		}

		if argsLen < 3 || argsLen > 4 {
			err := errors.New("wrong syntax: 'repoURL' 'fileURL' 'prefix' 'encryptMethod(default to none)'. prefix in appv1 means 'os/arch'")
			fmt.Println(err)
			return err
		}

		repoURL := context.Args().Get(0)
		fileURL := context.Args().Get(1)
		prefix := context.Args().Get(2)

		repo, err := client.NewUCRepo(repoURL)
		if err != nil {
			fmt.Println(err)
			return err
		}

		content, err := ioutil.ReadFile(fileURL)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = repo.Put(prefix+"/"+filepath.Base(fileURL), content, encrypt)
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
		//TODO: we can have a default server
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
		fmt.Println("success in downloading and verifing file: ", localFile)
		return nil
	},
}

var deleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a file from a repository",

	Action: func(context *cli.Context) error {
		if len(context.Args()) != 2 {
			err := errors.New("wrong syntax: pull 'repo url' 'filename'")
			fmt.Println(err)
			return err
		}

		repoURL := context.Args().Get(0)
		fileName := context.Args().Get(1)
		uc := new(cutils.UpdateClient)

		err := uc.Delete(repoURL, fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	},
}

var decryptCommand = cli.Command{
	Name:  "decrypt",
	Usage: "initiate default setting",
	Action: func(context *cli.Context) error {
		if len(context.Args()) != 2 {
			err := errors.New("wrong syntax: decrypt 'private key url' 'encrypted file url'")
			fmt.Println(err)
			return err
		}

		privFile := context.Args().Get(0)
		encryptedFile := context.Args().Get(1)
		privBytes, err := ioutil.ReadFile(privFile)
		if err != nil {
			fmt.Println(err)
			return err
		}
		encryptedBytes, err := ioutil.ReadFile(encryptedFile)
		if err != nil {
			fmt.Println(err)
			return err
		}

		decryptedBytes, err := utils.RSADecrypt(privBytes, encryptedBytes)
		if err != nil {
			fmt.Println(err)
			return err
		}

		decryptedFile := encryptedFile + "-decrypted"
		err = ioutil.WriteFile(decryptedFile, decryptedBytes, 0644)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("Success to decrypt %s to %s\n", encryptedFile, decryptedFile)
		return nil
	},
}
