package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/dockyard/cmd"
	_ "github.com/containerops/dockyard/middleware/notifications"
	"github.com/containerops/wrench/setting"
)

func main() {
	if err := setting.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read config failed: %v", err.Error())
		return
	}

	if err := setting.GetConfFromJSON("conf/config.json"); err != nil {
		fmt.Printf("Read middleware config failed and skip its function. %v", err.Error())
		//return
	}

	if err := drivers.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read backend config failed: %v", err.Error())
		return
	}

	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
