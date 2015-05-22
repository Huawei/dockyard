package middleware

import (
	"github.com/Unknwon/macaron"

	_ "github.com/macaron-contrib/session/redis"
)

func SetMiddlewares(m *macaron.Macaron) {
	//设置静态文件目录，静态文件的访问不进行日志输出
	m.Use(macaron.Static("static", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	//设置全局 Logger
	m.Map(Log)
	//设置 logger 的 Handler 函数，处理所有 Request 的日志输出
	m.Use(logger())

	//设置 panic 的 Recovery
	m.Use(macaron.Recovery())
}
