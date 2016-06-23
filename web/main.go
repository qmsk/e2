package web

import (
	"net/http"
	"log"
	"path"
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

func (options Options) RouteStatic(prefix string) Route {
	var route = Route{Pattern:prefix}

	if options.Static != "" {
		log.Printf("Serve static %v from %v\n", prefix, options.Static)

		route.Handler = http.StripPrefix(prefix, http.FileServer(http.Dir(options.Static)))
	}

	return route
}

// Return a route that serves a named static file on /
func (options Options) RouteDefaultFile(name string) Route {
	path := path.Join(options.Static, name)

	return Route{
		Pattern: "/",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				w.WriteHeader(404)
			} else {
				http.ServeFile(w, r, path)
			}
		}),
	}
}

func (options Options) Server(routes ...Route) {
	var serveMux = http.NewServeMux()

	for _, route := range routes {
		if route.Handler == nil {
			continue
		}

		serveMux.Handle(route.Pattern, route.Handler)
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
