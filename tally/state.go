package tally

import (
    "fmt"
)

// Tally ID
type ID int

type Output struct {
    Source      string
    Name        string
}

type Input struct {
    Source      string
    Name        string
}

type Status struct {
    Program     bool
    Preview     bool
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
    Input           Input
    Output          Output
    Tally			ID
    Status          Status
}

type State struct {
    Inputs  map[Input]ID
	Outputs map[Output]bool

    Links   []Link

    Tally   map[ID]Status
}

func makeState() State {
	return State{
        Inputs:  make(map[Input]ID),
		Outputs: make(map[Output]bool),
        Tally:   make(map[ID]Status),
    }
}

func (state *State) addInput(source string, name string, id ID) Input {
	input := Input{source, name}

	state.Inputs[input] = id

	return input
}

func (state *State) addOutput(source string, name string) Output {
	output := Output{source, name}

	state.Outputs[output] = true

	return output
}

func (state *State) addLink(link Link) {
    state.Links = append(state.Links, link)
}

// Update finaly Tally state from links
func (state *State) update() error {
    for _, link := range state.Links {
        status := state.Tally[link.Tally]

        if link.Status.Program {
            status.Program = true
        }
        if link.Status.Preview {
            status.Preview = true
        }

        state.Tally[link.Tally] = status
    }

    return nil
}

func (state State) Print() {
    fmt.Printf("Inputs: %d\n", len(state.Inputs))
    for input, id := range state.Inputs {
        fmt.Printf("\t%d: %s/%s\n", id, input.Source, input.Name)
    }
    fmt.Printf("Link: %d\n", len(state.Links))
    for _, link := range state.Links {
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
