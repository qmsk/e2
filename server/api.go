package server

import (
    "fmt"
    "net/http"
    "encoding/json"
    "strings"
)

type apiResource interface{}

type apiIndex interface {
    Index(name string) (apiResource, error)
}

type apiGET interface {
    Get() (interface{}, error)
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := strings.Split(r.URL.Path, "/")

    // lookup from root
    var resource apiResource

    resource = server

    for _, name := range path {
        if name == "" {
            continue
        }

        if indexResource, ok := resource.(apiIndex); !ok {
           w.WriteHeader(http.StatusNotFound)
        } else if nextResource, err := indexResource.Index(name); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "%v\n", err)
        } else {
            resource = nextResource
            continue
        }

        return
    }

    switch r.Method {
    case "GET":
        if apiGET, ok := resource.(apiGET); !ok {
            w.WriteHeader(http.StatusMethodNotAllowed)
        } else if ret, err := apiGET.Get(); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
        } else if ret == nil {
            w.WriteHeader(http.StatusNotFound)
        } else {
            json.NewEncoder(w).Encode(ret)
        }
    default:
        w.WriteHeader(http.StatusNotImplemented)
    }
}
