package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils/setting"
)

func DiscoveryACIHandler(ctx *macaron.Context, log *logs.BeeLogger) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	t, err := template.ParseFiles(models.TemplatePath)
	if err != nil {
		log.Error("[ACI API] Failed to parse template file: %v", err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(ctx.Resp, fmt.Sprintf("%v", err))
		return
	}

	err = t.Execute(ctx.Resp, models.TemplateDesc{
		Namespace:  namespace,
		Repository: repository,
		Domains:    setting.Domains,
		ListenMode: setting.ListenMode,
	})
	if err != nil {
		log.Error("[ACI API] Failed to respond: %v", err.Error())
		fmt.Fprintf(ctx.Resp, fmt.Sprintf("%v", err))
	}
}
