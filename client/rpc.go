package client

import (
    "bytes"
    "fmt"
    "net/http"
    "encoding/json"
    "log"
)

// JSON-RPC 2.0 HTTP variant used by E2

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
    httpRequest, err := http.NewRequest("POST", client.rpcURL.String(), &requestBuffer)
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
