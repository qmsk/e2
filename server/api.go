package server

import (
    "fmt"
    "net/http"
    "encoding/json"
    "strings"
)

type apiError struct {
    Status      int
    Err         error
}

func (err apiError) Error() string {
    if err.Err == nil {
        return fmt.Sprintf("HTTP %d", err.Status)
    } else {
        return fmt.Sprintf("%v", err.Err)
    }
}

type apiResource interface{}

// apiResource that supports sub-resources
type apiIndex interface {
    Index(name string) (apiResource, error)
}

// apiResource that supports GET
type apiGET interface {
    // Perform any independent post-processing + JSON encoding in the request handler goroutine.
    // Must be goroutine-safe!
    Get() (interface{}, error)
}

func (server *Server) apiLookup(path string) (apiResource, error) {
    // lookup from root
    var resource apiResource

    resource = server

    for _, name := range strings.Split(path, "/") {
        if indexResource, ok := resource.(apiIndex); !ok {
            return resource, apiError{http.StatusNotFound, nil}
        } else if nextResource, err := indexResource.Index(name); err != nil {
            return resource, err
        } else if nextResource == nil {
            return nil, apiError{http.StatusNotFound, nil}
        } else {
            resource = nextResource
        }
    }

    return resource, nil
}

func (server *Server) apiGet(path string) (apiGET, error) {
    if resource, err := server.apiLookup(path); err != nil {
        return nil, err
    } else if getResource, ok := resource.(apiGET); !ok {
        return nil, apiError{http.StatusMethodNotAllowed, nil}
    } else {
        return getResource, nil
    }
}

func (server *Server) apiHandler(w http.ResponseWriter, r *http.Request) error {
    path := r.URL.Path

    switch r.Method {
    case "GET":
        if getResource, err := server.apiGet(path); err != nil {
            return err
        } else if ret, err := getResource.Get(); err != nil {
            return err
        } else if ret == nil {
            return apiError{http.StatusNotFound, nil}
        } else {
            json.NewEncoder(w).Encode(ret)

            return nil
        }
    default:
        return apiError{http.StatusNotImplemented, nil}
    }
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := server.apiHandler(w, r); err == nil {

    } else if apiError, ok := err.(apiError); !ok {
        http.Error(w, err.Error(), 500)
    } else if apiError.Err != nil {
        http.Error(w, apiError.Err.Error(), apiError.Status)
    } else {
        http.Error(w, "", apiError.Status)
    }
}
