package server

type Index struct {
    Sources     map[string]Source       `json:"sources"`
    Screens     map[string]ScreenState  `json:"screens"`
}

func (index *Index) loadSources(sources *Sources) error {
    if err := sources.load(); err != nil {
        return err
    }

    index.Sources = sources.sourceMap

    return nil
}

func (index *Index) loadScreens(screens *Screens) error {
    if err := screens.load(); err != nil {
        return err
    }

    index.Screens = make(map[string]ScreenState)

    for _, screen := range screens.screenMap {
        if screenState, err := screen.state(screens.client); err != nil {
            return err
        } else {
            index.Screens[screenState.String()] = screenState
        }
    }

    return nil
}

func (index Index) Get() (interface{}, error) {
    return index, nil
}
