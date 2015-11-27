package handler

import (
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/wrench/setting"
)

// TBD: discovery template should be keep in line with ACI
func DiscoveryACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	repo := ctx.Params(":repository")

	content := setting.Domains + "/" + repo + " " + setting.ListenMode + "://" + setting.Domains
	result := "<meta name=\"ac-discovery\" content=\"" + content + "/image/" + repo + "-{version}-{os}-{arch}.{ext}\">\r\n"
	result += "<meta name=\"ac-discovery-pubkeys\" content=\"" + content + "/pubkeys/aci-pubkeys.gpg\">\r\n"

	return http.StatusOK, []byte(result)
}
