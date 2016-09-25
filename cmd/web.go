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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/macaron.v1"

	"github.com/containerops/configure"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/web"
)

var address string
var port int64

// webCmd is sub command which start/stop dockyard's REST API.
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web sub command start/stop dockyard's REST API Service.",
	Long:  ``,
	Run:   runWeb,
}

// init()
func init() {
	RootCmd.AddCommand(webCmd)

	webCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0", "http or https listen address.")
	webCmd.Flags().Int64VarP(&port, "port", "p", 80, "the port of http.")
}

func runWeb(cmd *cobra.Command, args []string) {
	m := macaron.New()

	//Set Macaron Web Middleware And Routers
	web.SetDockyardMacaron(m)

	listenMode := configure.GetString("listenmode")
	switch listenMode {
	case "http":
		listenaddr := fmt.Sprintf("%s:%d", address, port)
		if err := http.ListenAndServe(listenaddr, m); err != nil {
			fmt.Printf("Start Dockyard http service error: %v\n", err.Error())

		}
		break
	case "https":
		listenaddr := fmt.Sprintf("%s:443", address)
		server := &http.Server{Addr: listenaddr, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS10}, Handler: m}
		if err := server.ListenAndServeTLS(configure.GetString("httpscertfile"), configure.GetString("httpskeyfile")); err != nil {
			fmt.Printf("Start Dockyard https service error: %v\n", err.Error())

		}
		break
	case "unix":
		listenaddr := fmt.Sprintf("%s", address)
		if utils.IsFileExist(listenaddr) {
			os.Remove(listenaddr)
		}

		if listener, err := net.Listen("unix", listenaddr); err != nil {
			fmt.Printf("Start Dockyard unix socket error: %v\n", err.Error())

		} else {
			server := &http.Server{Handler: m}
			if err := server.Serve(listener); err != nil {
				fmt.Printf("Start Dockyard unix socket error: %v\n", err.Error())

			}
		}
		break
	default:
		break
	}
}
