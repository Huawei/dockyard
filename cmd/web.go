package cmd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/web"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

var CmdWeb = cli.Command{
	Name:        "web",
	Usage:       "start dockyard web service",
	Description: "dockyard is the module of handler docker and rkt image.",
	Action:      runWeb,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address",
			Value: "0.0.0.0",
			Usage: "web service listen ip, default is 0.0.0.0; if listen with Unix Socket, the value is sock file path.",
		},
		cli.IntFlag{
			Name:  "port",
			Value: 80,
			Usage: "web service listen at port 80; if run with https will be 443.",
		},
	},
}

func runWeb(c *cli.Context) {
	m := macaron.New()

	//Set Macaron Web Middleware And Routers
	web.SetDockyardMacaron(m)

	switch setting.ListenMode {
	case "http":
		listenaddr := fmt.Sprintf("%s:%d", c.String("address"), c.Int("port"))
		if err := http.ListenAndServe(listenaddr, m); err != nil {
			fmt.Printf("Start Dockyard http service error: %v", err.Error())
		}
		break
	case "https":
		listenaddr := fmt.Sprintf("%s:443", c.String("address"))
		server := &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
		if err := server.ListenAndServeTLS(setting.HttpsCertFile, setting.HttpsKeyFile); err != nil {
			fmt.Printf("Start Dockyard https service error: %v", err.Error())
		}
		break
	case "unix":
		listenaddr := fmt.Sprintf("%s", c.String("address"))
		if utils.IsFileExist(listenaddr) {
			os.Remove(listenaddr)
		}

		if listener, err := net.Listen("unix", listenaddr); err != nil {
			fmt.Printf("Start Dockyard unix socket error: %v", err.Error())
		} else {
			server := &http.Server{Handler: m}
			if err := server.Serve(listener); err != nil {
				fmt.Printf("Start Dockyard unix socket error: %v", err.Error())
			}
		}
		break
	default:
		break
	}
}
