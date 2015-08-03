package middleware

import (
	"strings"

	"github.com/Unknwon/macaron"

	"github.com/containerops/wrench/setting"
)

//TBD:codes as below should be updated when user config management is ready
func setRespHeaders() macaron.Handler {
	return func(ctx *macaron.Context) {
		if flag := strings.Contains(ctx.Req.RequestURI, "v1"); flag == true {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("X-Docker-Registry-Standalone", "True")       //Standalone
			ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.Version) //Version
			ctx.Resp.Header().Set("X-Docker-Registry-Config", setting.RunMode)  //Config
		} else {
			ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
			ctx.Resp.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		}
	}
}
