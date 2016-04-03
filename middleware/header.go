package middleware

import (
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/utils/setting"
)

func setRespHeaders() macaron.Handler {
	return func(ctx *macaron.Context) {
		if flag := strings.Contains(ctx.Req.RequestURI, "/v1/"); flag == true {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("X-Docker-Registry-Standalone", setting.Standalone)
			ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.RegistryVersion)
			ctx.Resp.Header().Set("X-Docker-Registry-Config", setting.RunMode)
			ctx.Resp.Header().Set("X-Docker-Endpoints", setting.Domains)
		} else if flag := strings.Contains(ctx.Req.RequestURI, "/v2/"); flag == true {
			//ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
			//ctx.Resp.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%v\"", setting.Domains))
			ctx.Resp.Header().Set("Docker-Distribution-Api-Version", setting.DistributionVersion)
		} else {
			//rkt header set
		}
	}
}
