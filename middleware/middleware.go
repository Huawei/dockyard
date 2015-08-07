package middleware

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/wrench/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	//Set static file directory,static file access without log output
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	InitLog(setting.RunMode, setting.LogPath)

	//Set global Logger
	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger(setting.RunMode))

	m.Use(getRespHeader())

	//Set the response header info
	m.Use(setRespHeaders())

	//Set recovery handler to returns a middleware that recovers from any panics
	m.Use(macaron.Recovery())
}
