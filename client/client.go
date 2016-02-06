package client

import (
    "fmt"
    "net/http"
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
}

func (client *Client) String() string {
    return client.url.String()
}
