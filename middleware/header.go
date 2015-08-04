package middleware

import (
	"fmt"
	"strings"

	"github.com/Unknwon/macaron"

	"github.com/containerops/wrench/setting"
)

//TBD:codes as below should be updated when user config management is ready
func setRespHeaders() macaron.Handler {
	return func(ctx *macaron.Context) {
		if flag := strings.Contains(ctx.Req.RequestURI, "v1"); flag == true {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("X-Docker-Registry-Standalone", setting.Standalone)   //Standalone
			ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.RegistryVersion) //Version
			ctx.Resp.Header().Set("X-Docker-Registry-Config", setting.RunMode)          //Config
		} else {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"\"", setting.Domains))
			ctx.Resp.Header().Set("Docker-Distribution-Api-Version", setting.DistributionVersion)
		}
	}
}
