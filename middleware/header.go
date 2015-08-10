package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Unknwon/macaron"

	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

var (
	ping = []string{"/v1/_ping", "/v1/_ping/", "/v2/", "/v2"}
)

func getRespHeader() macaron.Handler {
	return func(ctx *macaron.Context) {
		if flag, err := utils.Contain(ping, strings.Split(ctx.Req.RequestURI, "/")); err != nil {
			ctx.JSON(http.StatusBadRequest, "Docker registry or distribution's URL is invalid")
		} else if flag == true {

		}
	}
}

func setRespHeaders() macaron.Handler {
	return func(ctx *macaron.Context) {
		if flag, err := utils.Contain("v1", strings.Split(ctx.Req.RequestURI, "/")); err != nil {
			ctx.JSON(http.StatusBadRequest, "Docker registry or distribution's URL is invalid")
		} else if flag == true {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("X-Docker-Registry-Standalone", setting.Standalone)
			ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.RegistryVersion)
			ctx.Resp.Header().Set("X-Docker-Registry-Config", setting.RunMode)
			ctx.Resp.Header().Set("X-Docker-Endpoints", setting.Domains)
		} else if flag == false {
			ctx.Resp.Header().Set("Content-Type", "application/json")
			ctx.Resp.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%v\"", setting.Domains))
			ctx.Resp.Header().Set("Docker-Distribution-Api-Version", setting.DistributionVersion)
		}
	}
}
