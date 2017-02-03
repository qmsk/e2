package tally

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
)

func newSource(tally *Tally, discoveryPacket discovery.Packet, clientOptions client.Options) (Source, error) {
	source := Source{
		options:         tally.options,
		created:         time.Now(),
		discoveryPacket: discoveryPacket,
		clientOptions:   clientOptions,
	}

	// do not connect to slave VPs
	if discoveryPacket.MasterMac != discoveryPacket.MacAddress {
		return source, nil
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
	options Options
	created time.Time
	updated time.Time

	discoveryPacket discovery.Packet
	clientOptions   client.Options
	xmlClient       *client.XMLClient

	system client.System
	closed bool
	err    error
}

func (source Source) String() string {
	return source.clientOptions.String()
}

func (source Source) run(updateChan chan Source) {
	defer source.xmlClient.Close()
	for {
		if system, err := source.xmlClient.Read(); err == nil {
			source.system = system
		} else if err == io.EOF {
			source.closed = true
		} else {
			source.err = err
		}

		updateChan <- source

		if source.closed || source.err != nil {
			log.Printf("tally:Source %v: closed", source)
			return
		}
	}
}

func (source Source) updateState(state *State) error {
	tallySource := source.String()

	if source.err != nil {
		return source.err
	}

	if source.xmlClient == nil {
		// not connected
		return nil
	}

	system := source.system

	for sourceID, systemSource := range system.SrcMgr.SourceCol {
		// lookup Input from inputCfg with tally=
		if systemSource.InputCfgIndex < 0 {
			continue
		}
		inputCfg := system.SrcMgr.InputCfgCol[systemSource.InputCfgIndex]
		inputName := inputCfg.Name

		// resolve ID
		var tallyID ID

		if match := source.options.contactIDRegexp.FindStringSubmatch(inputCfg.ConfigContact); match == nil {
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
			// ignore?
			if source.options.ignoreDestRegexp != nil && source.options.ignoreDestRegexp.MatchString(screen.Name) {
				continue
			}

			var status Status

			if screen.IsActive > 0 {
				status.Active = true
			}

			if screen.Transition[0].InProgress() {
				status.Transition = screen.Transition[0].Progress()

				log.Printf("tally:Source %v: Screen %v: Transition in progress: %#v", source, screen.Name, status.Transition)
			}

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
			// ignore?
			if source.options.ignoreDestRegexp != nil && source.options.ignoreDestRegexp.MatchString(aux.Name) {
				continue
			}

			output := state.addOutput(tallySource, aux.Name)

			if aux.PvwLastSrcIndex == sourceID || aux.PgmLastSrcIndex == sourceID {
				state.addLink(Link{
					Tally:  tallyID,
					Input:  input,
					Output: output,
					Status: Status{
						Preview: (aux.PvwLastSrcIndex == sourceID),
						Program: (aux.PgmLastSrcIndex == sourceID),
						Active:  (aux.IsActive > 0),
					},
				})
			}
		}
	}

	return nil
}

// close, causing run() to exit
func (source Source) close() {
	if source.xmlClient != nil {
		source.xmlClient.Close()
	}
}

func (source Source) isClosed() bool {
	return source.xmlClient == nil || source.closed
}
