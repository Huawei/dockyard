package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/router"
)

func SetDockyardMacaron(m *macaron.Macaron) {
	//Setting Middleware
	middleware.SetMiddlewares(m)
	//Setting Router
	router.SetRouters(m)
}
