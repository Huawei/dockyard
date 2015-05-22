package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/setting"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/web"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "启动 dockyard 的 Web 服务",
	Description: "dockyard 提供 Docker 镜像仓库存储服务。",
	Action:      runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "Web 服务监听的 IP，默认 0.0.0.0；如果使用 Unix Socket 模式是 sock 文件的路径",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 80,
			Usage: "Web 服务监听的端口，默认 80",
		},
	},
}

func runWeb(c *cli.Context) {
	m := web.NewInstance()

	switch setting.ListenMode {
	case "http":
		listenaddr := fmt.Sprintf("%s:%d", c.String("address"), c.Int("port"))
		if err := http.ListenAndServe(listenaddr, m); err != nil {
			fmt.Printf("启动 dockyard 的 HTTP 服务错误: %v", err)
		}
		break
	case "https":
		//HTTPS 强制使用 443 端口
		listenaddr := fmt.Sprintf("%s:443", c.String("address"))
		server := &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
		if err := server.ListenAndServeTLS(setting.HttpsCertFile, setting.HttpsKeyFile); err != nil {
			fmt.Printf("启动 dockyard 的 HTTPS 服务错误: %v", err)
		}
		break
	case "unix":
		listenaddr := fmt.Sprintf("%s", c.String("address"))
		//如果存在 Unix Socket 文件就删除
		if utils.Exist(listenaddr) {
			os.Remove(listenaddr)
		}

		if listener, err := net.Listen("unix", listenaddr); err != nil {
			fmt.Printf("启动 dockyard 的 Unix Socket 监听错误: %v", err)
		} else {
			server := &http.Server{Handler: m}
			if err := server.Serve(listener); err != nil {
				fmt.Printf("启动 dockyard 的 Unix Socket 监听错误: %v", err)
			}
		}
		break
	default:
		break
	}
}
