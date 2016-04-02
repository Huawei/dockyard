package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/handler"
)

func SetRouters(m *macaron.Macaron) {
	m.Get("/", handler.IndexV1Handler)
}
