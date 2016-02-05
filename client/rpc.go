package client

import (
    "bytes"
    "fmt"
    "net/http"
    "encoding/json"
    "log"
    "time"
)

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

type Request struct {
    Method      string          `json:"method"`
    ID          int             `json:"id"`
    Params      interface{}     `json:"params"`
    Version     string          `json:"jsonrpc"`
}

type Response struct {
    ID          int             `json:"id"`
    Version     string          `json:"jsonrpc"`
    Result      interface{}     `json:"result"`
    Error       *Error          `json:"error"`
}

type Error struct {
    Code        int                 `json:"code"`
    Message     string              `json:"message"`
    Data        *json.RawMessage    `json:"data"`
}

func (err Error) Error() string {
    return err.Message
}

func (client *Client) do(request *Request, response *Response) error {
    client.seq++

    // encode
    var requestBuffer bytes.Buffer

    request.ID = client.seq
    request.Version = "2.0"

    if err := json.NewEncoder(&requestBuffer).Encode(request); err != nil {
        return err
    }

    // request
    httpRequest, err := http.NewRequest("POST", client.url.String(), &requestBuffer)
    if err != nil {
        return err
    }
    httpRequest.Header.Set("Content-Type", "application/json-rpc")

    log.Printf("%v %v: %v\n", httpRequest.Method, httpRequest.URL, requestBuffer.String())

    httpResponse, err := client.httpClient.Do(httpRequest)
    if err != nil {
        return err
    }

    // Decode
    defer httpResponse.Body.Close()

    switch contentType := httpResponse.Header.Get("Content-Type"); contentType {
    case "application/json", "application/json-rpc":
        if err := json.NewDecoder(httpResponse.Body).Decode(response); err != nil {
            return err
        }
    default:
        return fmt.Errorf("Invalid response: Content-Type=%#v", contentType)
    }

    if response.ID != request.ID {
        return fmt.Errorf("Response ID mismatch: request.ID=%v != response.ID=%v", request.ID, response.ID)
    }

    if response.Error != nil {
        return response.Error
    }

    return nil
}
