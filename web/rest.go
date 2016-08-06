package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Error struct {
	Status int
	Err    error
}

func (err Error) Error() string {
	if err.Err == nil {
		return fmt.Sprintf("HTTP %d", err.Status)
	} else {
		return fmt.Sprintf("%v", err.Err)
	}
}

type Resource interface{}

// Resource that supports sub-Resources
type IndexResource interface {
	Index(name string) (Resource, error)
}

// apiResource that supports GET
type GetResource interface {
	// Perform any independent post-processing + JSON encoding in the request handler goroutine.
	// Must be goroutine-safe!
	Get() (interface{}, error)
}

type API struct {
	root Resource
}

func MakeAPI(root Resource) API {
	return API{
		root: root,
	}
}

func (api API) index(path string) (Resource, error) {
	// lookup from root
	var resource = api.root

	for _, name := range strings.Split(path, "/") {
		if indexResource, ok := resource.(IndexResource); !ok {
			return resource, Error{http.StatusNotFound, nil}
		} else if nextResource, err := indexResource.Index(name); err != nil {
			return resource, err
		} else if nextResource == nil {
			return nil, Error{http.StatusNotFound, nil}
		} else {
			resource = nextResource
		}
	}

	return resource, nil
}

func (api API) handle(w http.ResponseWriter, r *http.Request) error {
	resource, err := api.index(r.URL.Path)

	if err != nil {
		return err
	}

	switch r.Method {
	case "GET":
		if getResource, ok := resource.(GetResource); !ok {
			return Error{http.StatusMethodNotAllowed, nil}
		} else if ret, err := getResource.Get(); err != nil {
			return err
		} else if ret == nil {
			return Error{http.StatusNotFound, nil}
		} else {
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(ret)

			return nil
		}
	default:
		return Error{http.StatusNotImplemented, nil}
	}
}

func (api API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := api.handle(w, r); err == nil {

	} else if httpError, ok := err.(Error); !ok {
		http.Error(w, err.Error(), 500)
	} else if httpError.Err != nil {
		http.Error(w, httpError.Err.Error(), httpError.Status)
	} else {
		http.Error(w, "", httpError.Status)
	}
}
