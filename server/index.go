package server

type Index struct {
    sources     map[string]Source
    Screens     map[string]ScreenState  `json:"screens"`

    // built in Get()
    Sources     map[string]SourceState  `json:"sources"`
}

func (index *Index) loadSources(sources *Sources) error {
    if err := sources.load(); err != nil {
        return err
    }

    index.sources = sources.sourceMap

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
    index.Sources = make(map[string]SourceState)

    for sourceName, source := range index.sources {
        index.Sources[sourceName] = source.buildState(index.Screens)
    }

    return index, nil
}
