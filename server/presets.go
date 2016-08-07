package server

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/web"
	"net/http"
)

type Presets struct {
	system *client.System
	jsonClient *client.JSONClient
	tcpClient *client.TCPClient

	presetMap map[string]Preset
}

func (presets *Presets) load() error {
	presetMap := make(map[string]Preset)

	for presetID, preset := range presets.system.PresetMgr.Preset {
		// parse sno
		preset := Preset{
			Preset:	preset,

			Group: preset.Sno.Group,
			Index: preset.Sno.Index,
		}

		presetMap[fmt.Sprintf("%v", presetID)] = preset
	}

	presets.presetMap = presetMap

	return nil
}

func (presets *Presets) Get() (interface{}, error) {
	return presets.presetMap, nil
}

func (presets *Presets) Post(request *http.Request) (interface{}, error) {
	var params struct {
		ID        int  `json:"id"`
		Live      bool `json:"live,omitempty"`
		Cut	      bool `json:"cut,omitempty"`
		AutoTrans int  `json:"autotrans,omitempty"`
	}
	params.ID = -1
	params.AutoTrans = -1

	if err := web.DecodeRequest(request, &params); err != nil {
		return nil, err
	}

	if params.ID >= 0 {
		preset , exists := presets.presetMap[fmt.Sprintf("%d", params.ID)]
		if !exists {
			return nil, nil // 404
		}

		if params.Live {
			if err := preset.take(presets.tcpClient); err != nil {
				return nil, err
			}
		} else {
			if err := preset.activate(presets.tcpClient); err != nil {
				return nil, err
			}
		}
	}

	// preview -> program
	if params.Cut {
		if err := presets.tcpClient.Cut(); err != nil {
			return nil, err
		}
	} else if params.AutoTrans == 0 {
		if err := presets.tcpClient.AutoTrans(); err != nil {
			return nil, err
		}
	} else if params.AutoTrans >= 0 {
		if err := presets.tcpClient.AutoTransFrames(params.AutoTrans); err != nil {
			return nil, err
		}
	}

	return params, nil
}

func (presets *Presets) Index(name string) (web.Resource, error) {
	if name == "" {
		var presetStates PresetStates

		for _, preset := range presets.presetMap {
			if presetState, err := preset.loadState(presets.jsonClient); err != nil {
				return nil, err
			} else {
				presetStates = append(presetStates, presetState)
			}
		}

		return presetStates, nil

	} else if preset, found := presets.presetMap[name]; !found {
		return nil, nil
	} else if presetState, err := preset.loadState(presets.jsonClient); err != nil {
		return presetState, err
	} else {
		return presetState, nil
	}
}

type PresetStates []PresetState

func (presetStates PresetStates) Get() (interface{}, error) {
	return presetStates, nil
}

type Preset struct {
	client.Preset

	// Decomposed Sequence order
	Group int `json:"group"`
	Index int `json:"index"`
}

func (preset Preset) String() string {
	return fmt.Sprintf("%d", preset.ID)
}

func (preset Preset) Get() (interface{}, error) {
	return preset, nil
}

func (preset Preset) activate(clientAPI client.API) error {
	if err := clientAPI.PresetRecall(preset.Preset); err != nil {
		return fmt.Errorf("RecallPreset %d: %v", preset.ID, err)
	}

	return nil
}

func (preset Preset) take(clientAPI client.API) error {
	if err := clientAPI.PresetAutoTrans(preset.Preset); err != nil {
		return fmt.Errorf("RecallPreset %d: %v", preset.ID, err)
	}

	return nil
}

func (preset Preset) loadState(jsonClient *client.JSONClient) (PresetState, error) {
	presetState := PresetState{Preset: preset}

	return presetState, presetState.load(jsonClient)
}

type PresetState struct {
	Preset

	Screens []string `json:"screens"`
	Auxes   []string `json:"auxes"`
}

func (presetState *PresetState) load(jsonClient *client.JSONClient) error {
	if presetDestinations, err := jsonClient.ListDestinationsForPreset(presetState.ID); err != nil {
		return err
	} else {
		for _, auxDest := range presetDestinations.AuxDest {
			auxID := fmt.Sprintf("%d", auxDest.ID)

			presetState.Auxes = append(presetState.Auxes, auxID)
		}

		for _, screenDest := range presetDestinations.ScreenDest {
			screenID := fmt.Sprintf("%d", screenDest.ID)

			presetState.Screens = append(presetState.Screens, screenID)
		}

		return nil
	}
}

func (presetState PresetState) Get() (interface{}, error) {
	return presetState, nil
}
