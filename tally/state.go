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
    Program         bool
    Preview         bool
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
    ID              ID
    Status          Status
}

type State struct {
    Inputs  map[Input]ID

    Links   []Link

    Tally   map[ID]Status
}

func (state *State) addLink(link Link) {
    state.Links = append(state.Links, link)
}

// Update finaly Tally state from links
func (state *State) update() error {
    for _, link := range state.Links {
        status := state.Tally[link.ID]

        if link.Status.Program {
            status.Program = true
        }
        if link.Status.Preview {
            status.Preview = true
        }

        state.Tally[link.ID] = status
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
        fmt.Printf("\t%d: %20s / %-15s <- %20s / %-15s = %v\n", link.ID,
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
