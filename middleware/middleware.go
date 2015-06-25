package middleware

import (
	"github.com/Unknwon/macaron"
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
	//Set the response header info
	m.Use(respHeaderSet())

	//Set recovery handler to returns a middleware that recovers from any panics
	m.Use(macaron.Recovery())
}
