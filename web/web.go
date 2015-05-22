package web

import (
	"github.com/Unknwon/macaron"

	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/router"
)

func NewInstance() *macaron.Macaron {
	m := macaron.New()

	//设置 Setting

	//设置 Middleware
	middleware.SetMiddlewares(m)
	//设置 Router
	router.SetRouters(m)

	return m
}
