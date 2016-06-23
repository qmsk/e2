package web

import (
	"net/http"
	"log"
)

type Options struct {
	Listen string `long:"http-listen" value-name:"[HOST]:PORT" default:":8284"`
	Static string `long:"http-static" value-name:"PATH"`
}

type Route struct {
	Pattern		string
	Handler		http.Handler
}

func RoutePrefix(prefix string, handler http.Handler) Route {
	return Route{
		Pattern: prefix,
		Handler: http.StripPrefix(prefix, handler),
	}
}

func (options Options) Server(routes ...Route) {
	var serveMux = http.NewServeMux()

	for _, route := range routes {
		serveMux.Handle(route.Pattern, route.Handler)
	}

	if options.Static != "" {
		log.Printf("Serve / from %v\n", options.Static)

		serveMux.Handle("/", http.FileServer(http.Dir(options.Static)))
	}

	if options.Listen != "" {
		var server = http.Server{
			Addr:		options.Listen,
			Handler:	serveMux,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Printf("http:Server.ListenAndServe %v: %v", options.Listen, err)
		}
	}
}
