package client

import (
	"fmt"
)

// E2 has it's own little response-in-result-in-response thing going on
type Result struct {
	Response interface{} `json:"response"`
	Success  int         `json:"success"`
}

func (client *Client) doResult(request *Request, data interface{}) error {
	result := Result{
		Response: data,
	}
	response := Response{
		Result: &result,
	}

	if err := client.do(request, &response); err != nil {
		return fmt.Errorf("RPC %v error: %v", request.Method, err)
	}

	//log.Printf("success=%v: %#v\n", result.Success, result.Response)

	// TODO: decode as json.RawMessage to decode error string..
	if result.Success != 0 {
		return fmt.Errorf("Nonzero response: success=%v", result.Success)
	}

	return nil
}
