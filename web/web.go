package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/router"
)

func SetMacaron(m *macaron.Macaron) {
	//设置 Setting

	//设置 Middleware
	middleware.SetMiddlewares(m)
	//设置 Router
	router.SetRouters(m)
}
