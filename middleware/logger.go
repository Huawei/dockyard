package middleware

import (
	"fmt"

	"github.com/Unknwon/macaron"

	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/setting"
)

var Log *logs.BeeLogger //全局日志对象

func init() {
	Log = logs.NewLogger(10000)

	if setting.RunMode == "dev" {
		Log.SetLogger("console", "")
	}

	Log.SetLogger("file", fmt.Sprintf("{\"filename\":\"%s\"}", setting.LogPath))

}

func logger() macaron.Handler {
	return func(ctx *macaron.Context) {
		//在调试阶段为了便于阅读控制台的信息，输出空行和分隔符区分多个访问的日志
		if setting.RunMode == "dev" {
			Log.Trace("")
			Log.Trace("----------------------------------------------------------------------------------")
		}
		//默认输出 Request 的 Method、 URI 和 Header 的信息
		Log.Trace("[%s] [%s]", ctx.Req.Method, ctx.Req.RequestURI)
		Log.Trace("[Header] %v", ctx.Req.Header)
	}
}
