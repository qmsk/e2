package tally

import (
	"github.com/qmsk/e2/web"
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
	Error		error
}

type restState struct {
	Inputs	[]restInput
	Tally	[]restTally
	Errors	[]restError
}

func (state State) Get() (interface{}, error) {
	var rs restState

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
		rs.Errors = append(rs.Errors, restError{Source:source, Error:err})
	}

	return rs, nil
}

func (tally *Tally) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return tally.get(), nil

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
			eventChan <- state
		}
	}()

	return web.MakeEvents(eventChan)
}
