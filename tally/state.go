package tally

import (
	"fmt"
	"time"

	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
)

// Tally ID
type ID int

type SourceState struct {
	Discovery discovery.Packet
	FirstSeen time.Time
	LastSeen  time.Time

	Connected bool
	Error     error
}

type Output struct {
	Source string
	Name   string
}

type Input struct {
	Source string
	Name   string
}

type InputState struct {
	ID     ID
	Status string
}

type Status struct {
	Program bool
	Preview bool
	Active  bool

	// Transitioning from Preview -> Program
	Transition client.TransitionProgress
}

// Consider tally as being in the high state
func (status Status) High() bool {
	return status.Program || status.Transition.InProgress()
}

func (status Status) String() string {
	if status.Program && status.Preview {
		return "program+preview"
	} else if status.Program {
		return "program"
	} else if status.Preview {
		return "preview"
	} else {
		return ""
	}
}

type Link struct {
	Tally  ID
	Input  Input
	Output Output
	Status Status
}

type TallyState struct {
	Inputs  map[Input]bool
	Outputs map[Output]Status
	Errors  []error

	Status Status
}

type State struct {
	Sources map[string]SourceState

	Inputs  map[Input]InputState
	Outputs map[Output]bool

	links []Link

	Tally  map[ID]TallyState
	Errors []error
}

func newState() *State {
	return &State{
		Sources: make(map[string]SourceState),
		Inputs:  make(map[Input]InputState),
		Outputs: make(map[Output]bool),
		Tally:   make(map[ID]TallyState),
	}
}

func (state *State) setSource(source Source) {
	state.setSourceError(source, nil)
}

func (state *State) setSourceError(source Source, err error) {
	sourceState := SourceState{
		Discovery: source.discoveryPacket,
		FirstSeen: source.created,
		LastSeen:  source.updated,
		Connected: (source.xmlClient != nil),
		Error:     err,
	}

	state.Sources[source.String()] = sourceState
}

func (state *State) addTallyError(id ID, input Input, err error) {
	tallyState := state.Tally[id]

	tallyState.Errors = append(tallyState.Errors, err)

	state.Tally[id] = tallyState
}

func (state *State) addInput(source string, name string, id ID, status string) Input {
	input := Input{source, name}

	state.Inputs[input] = InputState{
		ID:     id,
		Status: status,
	}

	return input
}

func (state *State) addOutput(source string, name string) Output {
	output := Output{source, name}

	state.Outputs[output] = true

	return output
}

func (state *State) addLink(link Link) {
	if _, exists := state.Inputs[link.Input]; !exists {
		panic(fmt.Errorf("addLink with unknown Input: %#v", link))
	}

	state.links = append(state.links, link)
}

// Update finaly Tally state from links
func (state *State) update() {
	for _, sourceState := range state.Sources {
		if sourceState.Error != nil {
			state.Errors = append(state.Errors, sourceState.Error)
		}
	}

	for input, inputState := range state.Inputs {
		tallyState := state.Tally[inputState.ID]

		if tallyState.Inputs == nil {
			tallyState.Inputs = make(map[Input]bool)
		}
		tallyState.Inputs[input] = true

		state.Tally[inputState.ID] = tallyState
	}

	for _, link := range state.links {
		tallyState := state.Tally[link.Tally]

		if tallyState.Outputs == nil {
			tallyState.Outputs = make(map[Output]Status)
		}
		tallyState.Outputs[link.Output] = link.Status

		if link.Status.Program {
			tallyState.Status.Program = true
		}
		if link.Status.Preview {
			tallyState.Status.Preview = true
		}
		if link.Status.Active {
			tallyState.Status.Active = true
		}
		if link.Status.Transition.InProgress() {
			// TODO: merge multiple transitions?
			tallyState.Status.Transition = link.Status.Transition
		}

		state.Tally[link.Tally] = tallyState
	}
}

func (state State) Print() {
	fmt.Printf("Inputs: %d\n", len(state.Inputs))
	for input, id := range state.Inputs {
		fmt.Printf("\t%d: %s/%s\n", id, input.Source, input.Name)
	}
	fmt.Printf("Link: %d\n", len(state.links))
	for _, link := range state.links {
		fmt.Printf("\t%d: %20s / %-15s <- %20s / %-15s = %v\n", link.Tally,
			link.Output.Source, link.Output.Name,
			link.Input.Source, link.Input.Name,
			link.Status,
		)
	}
	fmt.Printf("Tally: %d\n", len(state.Tally))
	for id, status := range state.Tally {
		fmt.Printf("\t%d: %v\n", id, status)
	}
	fmt.Printf("\n")
}
