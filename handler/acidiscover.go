package handler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/wrench/setting"
)

// TBD: discovery template should be updated to keep in line with ACI
func DiscoveryACIHandler(ctx *macaron.Context, log *logs.BeeLogger) {
	img := ctx.Params(":imagename")

	t, err := template.ParseFiles("conf/acifetchtemplate.html")
	if err != nil {
		log.Error("[ACI API] Discovery parse template file failed: %v", err.Error())
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(ctx.Resp, fmt.Sprintf("%v", err))
		return
	}

	err = t.Execute(ctx.Resp, struct {
		Name       string
		ServerName string
		Domain     string
	}{
		Name:       img,
		ServerName: setting.Domains,
		Domain:     setting.ListenMode,
	})
	if err != nil {
		log.Error("[ACI API] Discovery respond failed: %v", err.Error())
		fmt.Fprintf(ctx.Resp, fmt.Sprintf("%v", err))
	}
}
