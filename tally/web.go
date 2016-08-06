package tally

import (
	"github.com/qmsk/e2/web"
	"time"
)

type restInput struct {
	Input
	ID
	Status string
}

type restOutput struct {
	Output
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

type restSource struct {
	Source   string
	LastSeen string // duration
	Error    string `json:",omitempty"`

	SourceState
}

type restState struct {
	Sources []restSource
	Inputs  []restInput
	Outputs []restOutput
	Tally   []restTally
	Errors  []restError
}

type event struct {
	Tally *restState `json:"tally,omitempty"`
}

func (state State) toRest() (rs restState) {
	for sourceName, sourceState := range state.Sources {
		s := restSource{
			Source:      sourceName,
			SourceState: sourceState,
		}

		if !sourceState.LastSeen.IsZero() {
			s.LastSeen = time.Now().Sub(sourceState.LastSeen).String()
		}

		if sourceState.Error != nil {
			s.Error = sourceState.Error.Error()
		}

		rs.Sources = append(rs.Sources, s)
	}

	for input, inputState := range state.Inputs {
		rs.Inputs = append(rs.Inputs, restInput{Input: input, ID: inputState.ID, Status: inputState.Status})
	}

	for output, _ := range state.Outputs {
		rs.Outputs = append(rs.Outputs, restOutput{Output: output})
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

	for _, err := range state.Errors {
		rs.Errors = append(rs.Errors, restError{Error: err.Error()})
	}

	return
}

func (state State) Get() (interface{}, error) {
	return state.toRest(), nil
}

func (tally *Tally) Index(name string) (web.Resource, error) {
	switch name {
	case "tally":
		return tally.Get(), nil

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
