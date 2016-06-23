package tally

import (
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/web"
	"time"
)

type restInput struct {
	Input
	ID
}

type restStatus struct {
	Output
	Status
}

type restTally struct {
	ID			ID
	Inputs		[]Input
	Outputs		[]restStatus
	Status
}

type restError struct {
	Source		string
	Error		string
}

type restState struct {
	Inputs	[]restInput
	Tally	[]restTally
	Errors	[]restError
}

type restSource struct {
	Source	  string
	Discovery discovery.Packet
	FirstSeen time.Time
	LastSeen  string

	Error	  string	`json:",omitempty"`
}

type event struct {
	Tally	*restState `json:"tally,omitempty"`
}

func (sources sources) Get() (interface{}, error) {
	var rss []restSource

	for sourceName, source := range sources {
		var rs = restSource{
			Source:	   sourceName,
			Discovery: source.discoveryPacket,
			FirstSeen: source.created,
		}

		if !source.updated.IsZero() {
			rs.LastSeen = time.Now().Sub(source.updated).String()
		}

		if source.err != nil {
			rs.Error = source.err.Error()
		}

		rss = append(rss, rs)
	}

	return rss, nil
}

func (state State) toRest() (rs restState) {
	for input, id := range state.Inputs {
		rs.Inputs = append(rs.Inputs, restInput{Input:input, ID:id})
	}

	for id, tallyState := range state.Tally {
		var tally = restTally{
			ID:	id,
			Status: tallyState.Status,
		}

		for input, _ := range tallyState.Inputs {
			tally.Inputs = append(tally.Inputs, input)
		}

		for output, status := range tallyState.Outputs {
			tally.Outputs = append(tally.Outputs, restStatus{Output: output, Status: status })
		}


		rs.Tally = append(rs.Tally, tally)
	}

	for source, err := range state.Errors {
		rs.Errors = append(rs.Errors, restError{Source:source, Error:err.Error()})
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

	tally.register(stateChan)

	go func(){
		for state := range stateChan {
			restState := state.toRest()

			var event = event{Tally: &restState}

			eventChan <- event
		}
	}()

	return web.MakeEvents(eventChan)
}
