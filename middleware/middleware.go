package middleware

import (
	"fmt"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/utils/setting"
)

func SetMiddlewares(m *macaron.Macaron) {
	//Set static file directory,static file access without log output
	m.Use(macaron.Static("external", macaron.StaticOptions{
		Expires: func() string { return "max-age=0" },
	}))

	InitLog(setting.RunMode, setting.LogPath)

	//Set global Logger
	m.Map(Log)
	//Set logger handler function, deal with all the Request log output
	m.Use(logger(setting.RunMode))

	//Set the response header info
	m.Use(setRespHeaders())

	m.Use(Handlefunc())

	//Set recovery handler to returns a middleware that recovers from any panics
	m.Use(macaron.Recovery())
}

type HandlerInterface interface {
	InitFunc() error
	Handler(ctx *macaron.Context)
}

var Middleware map[string]HandlerInterface = map[string]HandlerInterface{}

func Register(name string, handler HandlerInterface) error {

	if _, existed := Middleware[name]; existed {
		return fmt.Errorf("%v has already been registered", name)
	}

	Middleware[name] = handler

	return nil
}

func Initfunc() error {
	var namespace []string = []string{setting.JSONConfCtx.Authors.Name(), setting.JSONConfCtx.Notifications.Name}

	for _, name := range namespace {
		if handlerinterface, existed := Middleware[name]; existed {
			if err := handlerinterface.InitFunc(); err != nil {
				return fmt.Errorf("Init %v failed, err: %v", name, err.Error())
			}
		}
	}

	return nil
}

func Handlefunc() macaron.Handler {
	return func(ctx *macaron.Context) {
		var namespace []string = []string{setting.JSONConfCtx.Authors.Name(), setting.JSONConfCtx.Notifications.Name}

		for _, name := range namespace {
			if handlerinterface, existed := Middleware[name]; existed {
				handlerinterface.Handler(ctx)
			}
		}
	}
}
