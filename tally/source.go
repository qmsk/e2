package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"io"
	"regexp"
	"time"
)

var INPUT_CONTACT_REGEXP = regexp.MustCompile("tally=(\\d+)")

func newSource(tally *Tally, discoveryPacket discovery.Packet, clientOptions client.Options) (Source, error) {
	source := Source{
		created:		 time.Now(),
		discoveryPacket: discoveryPacket,
		clientOptions:   clientOptions,
	}

	// give updates every 10s when idle
	clientOptions.ReadKeepalive = true

	if xmlClient, err := clientOptions.XMLClient(); err != nil {
		return source, err
	} else {
		source.xmlClient = xmlClient
	}

	go source.run(tally.sourceChan)

	return source, nil
}

// E2 XML source
//
// A source can either be in a running state with err == nil, or in a failed state with err != nil
type Source struct {
	created			time.Time
	updated			time.Time

	discoveryPacket	discovery.Packet
	clientOptions	client.Options
	xmlClient       *client.XMLClient

	system client.System
	closed bool
	err    error
}

func (source Source) String() string {
	return source.clientOptions.String()
}

func (source Source) run(updateChan chan Source) {
	for {
		if system, err := source.xmlClient.Read(); err == nil {
			source.system = system
		} else if err == io.EOF {
			source.closed = true
		} else {
			source.err = err
		}

		updateChan <- source

		if source.err != nil {
			break
		}
	}
}

func (source Source) updateState(state *State) error {
	tallySource := source.String()

	if source.err != nil {
		return source.err
	}

	system := source.system

	for sourceID, source := range system.SrcMgr.SourceCol.List() {
		// lookup Input from inputCfg with tally=
		if source.InputCfgIndex < 0 {
			continue
		}
		inputCfg := system.SrcMgr.InputCfgCol[source.InputCfgIndex]
		inputName := inputCfg.Name

		// resolve ID
		var tallyID ID

		if match := INPUT_CONTACT_REGEXP.FindStringSubmatch(inputCfg.ConfigContact); match == nil {
			continue
		} else if _, err := fmt.Sscanf(match[1], "%d", &tallyID); err != nil {
			return fmt.Errorf("Invalid Input Contact=%v: %v\n", inputCfg.ConfigContact, err)
		}

		input := state.addInput(tallySource, inputName, tallyID, inputCfg.InputCfgVideoStatus.String())

		// input state
		if inputCfg.InputCfgVideoStatus == client.InputVideoStatusBad {
			state.addTallyError(tallyID, input, fmt.Errorf("Source %v Input %v video status: %v", tallySource, inputName, inputCfg.InputCfgVideoStatus))
		}

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

			output := state.addOutput(tallySource, screen.Name)

			if status.Preview || status.Program {
				state.addLink(Link{
					Tally:  tallyID,
					Input:  input,
					Output: output,
					Status: status,
				})
			}
		}

		for _, aux := range system.DestMgr.AuxDestCol {
			output := state.addOutput(tallySource, aux.Name)

			if aux.PvwLastSrcIndex == sourceID || aux.PgmLastSrcIndex == sourceID {
				state.addLink(Link{
					Tally:  tallyID,
					Input:  input,
					Output: output,
					Status: Status{
						Preview: (aux.PvwLastSrcIndex == sourceID),
						Program: (aux.PgmLastSrcIndex == sourceID),
					},
				})
			}
		}
	}

	return nil
}

// close, causing run() to exit
func (source Source) close() {
	source.xmlClient.Close()
}
