package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/dockyard/cmd"
	"github.com/containerops/wrench/setting"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if err := setting.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read config error: %v", err.Error())
		return
	}

	if err := drivers.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read backend config error: %v", err.Error())
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
