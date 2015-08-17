package middleware

import (
	"fmt"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

var Log *logs.BeeLogger

func InitLog(runmode, path string) {
	Log = logs.NewLogger(10000)

	if runmode == "dev" {
		Log.SetLogger("console", "")
	}

	Log.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s\"}", path))

}

func logger(runmode string) macaron.Handler {
	return func(ctx *macaron.Context) {
		if runmode == "dev" {
			Log.Trace("")
			Log.Trace("----------------------------------------------------------------------------------")
		}

		Log.Trace("[%s] [%s]", ctx.Req.Method, ctx.Req.RequestURI)
		Log.Trace("[Header] %v", ctx.Req.Header)
	}
}
