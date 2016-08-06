package tally

import (
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/web"
	"time"
)

type restInput struct {
	Input
	ID
	Status string
}

type restStatus struct {
	Output
	Status
}

type restTally struct {
	ID      ID
	Inputs  []restInput
	Outputs []restStatus
	Status
	Errors []string
}

type restError struct {
	Source string
	Error  string
}

type restState struct {
	Inputs []restInput
	Tally  []restTally
	Errors []restError
}

type restSource struct {
	Source    string
	Discovery discovery.Packet
	FirstSeen time.Time
	LastSeen  string

	Connected bool
	Error     string `json:",omitempty"`
}

type event struct {
	Tally *restState `json:"tally,omitempty"`
}

func (sources sources) Get() (interface{}, error) {
	var rss []restSource

	for sourceName, source := range sources {
		var rs = restSource{
			Source:    sourceName,
			Discovery: source.discoveryPacket,
			FirstSeen: source.created,
		}

		if !source.updated.IsZero() {
			rs.LastSeen = time.Now().Sub(source.updated).String()
		}

		if source.xmlClient != nil {
			rs.Connected = true
		}

		if source.err != nil {
			rs.Error = source.err.Error()
		}

		rss = append(rss, rs)
	}

	return rss, nil
}

func (state State) toRest() (rs restState) {
	for input, inputState := range state.Inputs {
		rs.Inputs = append(rs.Inputs, restInput{Input: input, ID: inputState.ID, Status: inputState.Status})
	}

	for id, tallyState := range state.Tally {
		var tally = restTally{
			ID:     id,
			Status: tallyState.Status,
		}

		for input, _ := range tallyState.Inputs {
			inputState := state.Inputs[input]

			tally.Inputs = append(tally.Inputs, restInput{Input: input, ID: id, Status: inputState.Status})
		}

		for output, status := range tallyState.Outputs {
			tally.Outputs = append(tally.Outputs, restStatus{Output: output, Status: status})
		}

		for _, err := range tallyState.Errors {
			tally.Errors = append(tally.Errors, err.Error())
		}

		rs.Tally = append(rs.Tally, tally)
	}

	for source, err := range state.Errors {
		rs.Errors = append(rs.Errors, restError{Source: source, Error: err.Error()})
	}

	return
}

func (state State) Get() (interface{}, error) {
	return state.toRest(), nil
}

func (tally *Tally) Index(name string) (web.Resource, error) {
	switch name {
	case "sources":
		return tally.getSources(), nil

	case "tally":
		return tally.getState(), nil

	default:
		return nil, nil
	}
}

func (tally *Tally) WebAPI() web.API {
	return web.MakeAPI(tally)
}

func (tally *Tally) WebEvents() *web.Events {
	stateChan := make(chan State)
	eventChan := make(chan web.Event)

	tally.Register(stateChan)

	go func() {
		for state := range stateChan {
			restState := state.toRest()

			var event = event{Tally: &restState}

			eventChan <- event
		}
	}()

	return web.MakeEvents(eventChan)
}
