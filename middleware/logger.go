package middleware

import (
	"fmt"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/setting"
)

var Log *logs.BeeLogger

func init() {
	Log = logs.NewLogger(10000)

	if setting.RunMode == "dev" {
		Log.SetLogger("console", "")
	}

	Log.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s\"}", setting.LogPath))

}

func logger() macaron.Handler {
	return func(ctx *macaron.Context) {
		if setting.RunMode == "dev" {
			Log.Trace("")
			Log.Trace("----------------------------------------------------------------------------------")
		}
		Log.Trace("[%s] [%s]", ctx.Req.Method, ctx.Req.RequestURI)
		Log.Trace("[Header] %v", ctx.Req.Header)
	}
}
