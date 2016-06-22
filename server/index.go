package server

import (
	"github.com/qmsk/e2/client"
)

type Index struct {
	sources Sources
	screens Screens

	Screens map[string]ScreenState `json:"screens"`

	// built in Get()
	Sources map[string]SourceState `json:"sources"`
}

func (index *Index) load(client *client.Client) error {
	if err := index.sources.load(client); err != nil {
		return err
	}

	if err := index.screens.load(client); err != nil {
		return err
	}

	index.Screens = make(map[string]ScreenState)

	for _, screen := range index.screens.screenMap {
		if screenState, err := screen.loadState(client); err != nil {
			return err
		} else {
			index.Screens[screenState.String()] = screenState
		}
	}

	return nil
}

func (index Index) Get() (interface{}, error) {
	index.Sources = make(map[string]SourceState)

	for sourceName, source := range index.sources.sourceMap {
		index.Sources[sourceName] = source.buildState(index.Screens)
	}

	return index, nil
}
