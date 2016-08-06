package server

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/web"
)

type Presets struct {
	system *client.System
	jsonClient *client.JSONClient

	presetMap map[string]Preset
}

func (presets *Presets) load() error {
	presetMap := make(map[string]Preset)

	for presetID, preset := range presets.system.PresetMgr.Preset {
		// parse sno
		preset := Preset{
			Preset:	preset,
		}

		preset.Group, preset.Index = preset.ParseOrder()

		presetMap[fmt.Sprintf("%v", presetID)] = preset
	}

	presets.presetMap = presetMap

	return nil
}

func (presets *Presets) Get() (interface{}, error) {
	return presets.presetMap, nil
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
