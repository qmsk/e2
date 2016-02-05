package client

import (
    "fmt"
    "net/http"
    "log"
    "time"
)

type NotFound struct {
    ID      int
}

func (err NotFound) Error() string {
    return fmt.Sprintf("Not Found: %v", err.ID)
}

type Options struct {
    URL     URL             `long:"e2-jsonrpc-url" value-name:"http://IP:9999/"`
    Timeout time.Duration   `long:"e2-jsonrpc-timeout" default:"10s"`
}

func (options Options) Client() (*Client, error) {
    if options.URL.Empty() {
        return nil, fmt.Errorf("No URL given")
    }

    client := &Client{
        url:        options.URL,
        httpClient: &http.Client{
            Timeout:    options.Timeout,
        },
    }

    return client, nil
}

type Client struct {
    url             URL
    httpClient      *http.Client
    seq             int

    // internal state
    sources         map[int]Source
}

func (client *Client) updateNew(class string, id int, value interface{}) {
    log.Printf("New %s %d: %v\n", class, id, value)
}

func (client *Client) updateChange(class string, id int, value interface{}) {
    log.Printf("Update %s %d: %v\n", class, id, value)
}

func (client *Client) updateRemove(class string, id int, value interface{}) {
    log.Printf("Remove %s %d: %v\n", class, id, value)
}

func (client *Client) updateSources() error {
    // TODO: invalidate
    if client.sources != nil {
        return nil
    }

    // list
    sources, err := client.ListSources()
    if err != nil {
        return err
    }

    // map
    sourceMap := make(map[int]Source)

    for _, source := range sources {
        sourceMap[source.ID] = source
    }

    // diff
    for sourceID, source := range sourceMap {
        if prev, exists := client.sources[sourceID]; !exists {
            client.updateNew("source", sourceID, source)
        } else if source != prev {
            client.updateChange("source", sourceID, source)
        }
    }

    for sourceID, source := range client.sources {
        if _, exists := sourceMap[sourceID]; !exists {
            client.updateRemove("source", sourceID, source)
        }
    }

    client.sources = sourceMap

    return nil
}

func (client *Client) Sources() (sources []Source, err error) {
    if err := client.updateSources(); err != nil {
        return nil, err
    }

    for _, source := range client.sources {
        sources = append(sources, source)
    }

    return
}

func (client *Client) Source(id int) (Source, error) {
    if err := client.updateSources(); err != nil {
        return Source{}, err
    } else if source, exists := client.sources[id]; !exists {
        return source, NotFound{id}
    } else {
        return source, nil
    }
}
