package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdBackend = cli.Command{
	Name:        "backend",
	Usage:       "处理 dockyard 的后端存储服务",
	Description: "dockyard 支持使用一个或多个存储服务, 国内服务支持七牛、又拍、阿里云和腾讯云，国外服务支持亚马逊和谷歌云服务。",
	Action:      runBackend,
	Flags:       []cli.Flag{},
}

func runBackend(c *cli.Context) {

}
