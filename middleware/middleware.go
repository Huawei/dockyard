package middleware

import (
	"fmt"
	"github.com/Unknwon/macaron"
	"github.com/containerops/crew/setting"
	_ "github.com/macaron-contrib/session/redis"
)

func SetMiddlewares(m *macaron.Macaron) {
	//Set static file directory,static file access without log output
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	//Set global Logger
	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger())

	//TBD:codes as below should be updated when user config management is ready
	m.Use(func(ctx *macaron.Context) {
		ctx.Resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
		ctx.Resp.Header().Set("X-Docker-Registry-Standalone", "True")                                         //Standalone
		ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.Version)                                   //Version
		ctx.Resp.Header().Set("X-Docker-Registry-Config", "dev")                                              //Config
		ctx.Resp.Header().Set("X-Docker-Encrypt", "false")                                                    //Encrypt
		ctx.Resp.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\",Token", "containerops.me")) //docker V2
		ctx.Resp.Header().Set("Docker-Distribution-API-Version", "registry/2.0")                              //docker V2
	})

	//Set recovery handler to returns a middleware that recovers from any panics
	m.Use(macaron.Recovery())
}
