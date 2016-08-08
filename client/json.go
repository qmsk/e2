package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

// JSON-RPC 2.0 HTTP variant used by E2

type Request struct {
	Method  string      `json:"method"`
	ID      int         `json:"id"`
	Params  interface{} `json:"params"`
	Version string      `json:"jsonrpc"`
}

type Response struct {
	ID      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
}

type Error struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data"`
}

func (err Error) Error() string {
	return err.Message
}

// Returned errors
type NotFound struct {
	ID int
}

func (err NotFound) Error() string {
	return fmt.Sprintf("Not Found: %v", err.ID)
}

// E2 has it's own little response-in-result-in-response thing going on
type Result struct {
	Response interface{} `json:"response"`
	Success  int         `json:"success"`
}

func (options Options) JSONClient() (*JSONClient, error) {
	if options.Address == "" {
		return nil, fmt.Errorf("No Address given")
	}

	jsonClient := &JSONClient{
		options: options,
		rpcURL: url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(options.Address, options.JSONPort),
		},
		httpClient: &http.Client{
			Timeout: options.Timeout,
		},
	}

	return jsonClient, nil
}

type JSONClient struct {
	options Options

	// JSON RPC
	rpcURL     url.URL
	httpClient *http.Client
	seq        int
}

func (jsonClient *JSONClient) String() string {
	return jsonClient.options.Address
}

func (client *JSONClient) requestAPI(request *Request, response *Response) error {
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
			return fmt.Errorf("Invalid response JSON: %v", err)
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

func (client *JSONClient) request(request *Request, data interface{}) error {
	result := Result{
		Response: data,
	}
	response := Response{
		Result: &result,
	}

	if err := client.requestAPI(request, &response); err != nil {
		return fmt.Errorf("RPC %v error: %v", request.Method, err)
	}

	//log.Printf("success=%v: %#v\n", result.Success, result.Response)

	// TODO: decode as json.RawMessage to decode error string..
	if result.Success != 0 {
		return fmt.Errorf("Nonzero response: success=%v", result.Success)
	}

	return nil
}

func (client *JSONClient) readRequest(request *Request, data interface{}) error {
	return client.request(request, data)
}

func (client *JSONClient) safeRequest(request *Request, data interface{}) error {
	if client.options.Safe {
		return nil
	} else {
		return client.request(request, data)
	}
}

func (client *JSONClient) liveRequest(request *Request, data interface{}) error {
	if client.options.ReadOnly || client.options.Safe {
		return nil
	} else {
		return client.request(request, data)
	}
}


// Presets
const listPresetsExclude = -2
const listPresetsInclude = -1

type listPresets struct {
	ScreenDest int `json:"ScreenDest"`
	AuxDest    int `json:"AuxDest"`
}

// Default is to return all presets
func (client *JSONClient) ListPresets() (presetList []Preset, err error) {
	request := Request{
		Method: "listPresets",
		Params: struct{}{},
	}

	if err := client.readRequest(&request, &presetList); err != nil {
		return nil, err
	} else {
		return presetList, nil
	}
}

func (client *JSONClient) ListPresetsX(screenID int, auxID int) (presetList []Preset, err error) {
	request := Request{
		Method: "listPresets",
		Params: listPresets{
			ScreenDest: screenID,
			AuxDest:    auxID,
		},
	}

	if err := client.readRequest(&request, &presetList); err != nil {
		return nil, err
	} else {
		return presetList, nil
	}
}

// Presets
type PresetAuxDest struct {
	ID int `json:"id"`
}
type PresetScreenDest struct {
	ID int `json:"id"`
}

type PresetDestinations struct {
	Preset

	AuxDest    []PresetAuxDest    `json:"AuxDest"`
	ScreenDest []PresetScreenDest `json:"ScreenDest"`
}

type listDestinationsForPreset struct {
	ID int `json:"id"`
}

func (client *JSONClient) ListDestinationsForPreset(presetID int) (result PresetDestinations, err error) {
	if presetID < 0 {
		return result, fmt.Errorf("Invalid Preset ID: %v", presetID)
	}

	request := Request{
		Method: "listDestinationsForPreset",
		Params: listDestinationsForPreset{
			ID: presetID,
		},
	}

	if err := client.readRequest(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

type activatePreset struct {
	ID		int	`json:"id"`
	Type	int	`json:"type"`
}

func (client *JSONClient) activatePreset(id int, recallType int) error {
	if id < 0 {
		return fmt.Errorf("Invalid Preset ID: %v", id)
	}

	request := Request{
		Method: "activatePreset",
		Params: activatePreset{
			ID: id,
			Type: recallType,
		},
	}

	if recallType > 0 {
		// type == 1 -> program
		return client.liveRequest(&request, nil)
	} else {
		// type == 0 -> preview
		return client.safeRequest(&request, nil)
	}
}

func (client *JSONClient) ActivatePresetPreview(id int) error {
	return client.activatePreset(id, 0)
}
func (client *JSONClient) ActivatePresetProgram(id int) error {
	return client.activatePreset(id, 1)
}

// Destinations
type listDestinations struct {
	Type int `json:"type"`
}

const listDestinationsTypeAll = 0
const listDestinationsTypeScreen = 1
const listDestinationsTypeAux = 2

type ListDestinations struct {
	AuxDestinations    []AuxDest           `json:"AuxDestination"`
	ScreenDestinations []ScreenDestination `json:"ScreenDestination"`
}

func (client *JSONClient) ListDestinations() (result ListDestinations, err error) {
	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeAll,
		},
	}

	if err := client.readRequest(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

func (client *JSONClient) ListAuxDestinations() ([]AuxDest, error) {
	var result ListDestinations

	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeAux,
		},
	}

	if err := client.readRequest(&request, &result); err != nil {
		return nil, err
	} else {
		return result.AuxDestinations, nil
	}
}

func (client *JSONClient) ListScreenDestinations() ([]ScreenDestination, error) {
	var result ListDestinations

	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeScreen,
		},
	}

	if err := client.readRequest(&request, &result); err != nil {
		return nil, err
	} else {
		return result.ScreenDestinations, nil
	}
}

// Screen Content
type listContent struct {
	ID int `json:"id"`
}

type ListContent struct {
	ID   int    `json:"id"`
	Name string `json:"Name"`

	Layers   []*Layer  `json:"Layers"`
	BGLayers []BGLyr   `json:"BgLyr"`

	// Transition
}

func (client *JSONClient) ListContent(screenID int) (result ListContent, err error) {
	request := Request{
		Method: "listContent",
		Params: listContent{
			ID: screenID,
		},
	}

	if err := client.readRequest(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

// Sources
type listSources struct {
	Type int `json:"type"`
}

const listSourcesTypeInput = 0
const listSourcesTypeBackground = 1

// Default is to return all
func (client *JSONClient) ListSources() (sourceList []Source, err error) {
	request := Request{
		Method: "listSources",
		Params: struct{}{},
	}

	if err := client.readRequest(&request, &sourceList); err != nil {
		return nil, err
	} else {
		return sourceList, nil
	}
}
