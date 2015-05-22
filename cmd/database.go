package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdDatabase = cli.Command{
	Name:        "db",
	Usage:       "处理 dockyard 程序的数据库创建、备份和恢复等数据库维护",
	Description: "dockyard 使用 RebornDB 和 Redis 处理持久化数据",
	Action:      runDatabase,
	Flags:       []cli.Flag{},
}

func runDatabase(c *cli.Context) {}
