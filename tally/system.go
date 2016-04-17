package tally

import (
    "github.com/qmsk/e2/client"
    "fmt"
    "regexp"
)

var inputContactRegexp = regexp.MustCompile("tally=(\\d+)")

func (state *State) updateSystem(system client.System, tallySource string) error {
    for sourceID, source := range system.SrcMgr.SourceCol.List() {
        // lookup Input from inputCfg with tally= 
        if source.InputCfgIndex < 0 {
            continue
        }
        inputCfg := system.SrcMgr.InputCfgCol[source.InputCfgIndex]

        var tallyID ID

        if match := inputContactRegexp.FindStringSubmatch(inputCfg.ConfigContact); match == nil {
            continue
        } else if _, err := fmt.Sscanf(match[1], "%d", &tallyID); err != nil {
            return fmt.Errorf("Invalid Input Contact=%v: %v\n", inputCfg.ConfigContact, err)
        }

        input := Input{tallySource, inputCfg.Name}

        state.Inputs[input] = tallyID

        // lookup active Links
        for _, screen := range system.DestMgr.ScreenDestCol {
            var status Status

            for _, layer := range screen.LayerCollection {
                if layer.LastSrcIdx == sourceID {
                    if layer.PvwMode > 0 {
                        status.Preview = true
                    }
                    if layer.PgmMode > 0 {
                        status.Program = true
                    }
                }
            }
            if status.Preview || status.Program {
                state.addLink(Link{
                    Input:          input,
                    Output:         Output{tallySource, screen.Name},
                    ID:             tallyID,
                    Status:         status,
                })
            }
        }

        for _, aux := range system.DestMgr.AuxDestCol {
            var output = Output{tallySource, aux.Name}

            if aux.PvwLastSrcIndex == sourceID || aux.PgmLastSrcIndex == sourceID {
                state.addLink(Link{
                    Input:          input,
                    Output:         output,
                    ID:             tallyID,
                    Status: Status{
                        Preview:    (aux.PvwLastSrcIndex == sourceID),
                        Program:    (aux.PgmLastSrcIndex == sourceID),
                    },
                })
            }
        }
    }

    return nil
}
