package tally

import (
    "github.com/qmsk/e2/client"
)

func newSource(tally *Tally, clientOptions client.Options) (Source, error) {
    source := Source{
        clientOptions:  clientOptions,
    }

    if xmlClient, err := clientOptions.XMLClient(); err != nil {
        return source, err
    } else {
        source.xmlClient = xmlClient
    }

    go source.run(tally.sourceChan)

    return source, nil
}

type Source struct {
    clientOptions   client.Options
    xmlClient   *client.XMLClient

    system      client.System
    err         error
}

func (source Source) String() string {
    return source.clientOptions.String()
}

func (source Source) run(updateChan chan Source) {
    for {
        if system, err := source.xmlClient.Read(); err != nil {
            source.err = err
        } else {
            source.system = system
        }

        updateChan <- source

        if source.err != nil {
            break
        }
    }
}
