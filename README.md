# qmsk-e2
E2 Client + REST + WebSocket Server + Web UI

## Server

    go get ./cmd/server
    
    cd static && bower install

Follow E2 status, providing a REST + WebSocket API, and a web UI:

    $GOPATH/bin/server --discovery-interface=eth0 --http-listen=:8284 --http-static=./static

### API

TODO: examples

#### *GET* `/api/`

Combines both sources and screens, including cross-correlated program/preview state for both. This is `O(N)` RPCs on the number of screen destinations.

      {
        "sources": {
          "4" : {
             "dimensions" : {
                "width" : 1920,
                "height" : 1080
             },
             "type" : "input",
             "id" : 4,
             "name" : "PC 3",
             "status" : "ok"
          },
          "5" : {
             "id" : 5,
             "name" : "PC 4",
             "status" : "ok",
             "type" : "input",
             "dimensions" : {
                "height" : 1080,
                "width" : 1920
             }
          },
        },
        "screens" : {
          "0" : {
             "id" : 0,
             "preview_sources" : [
                "4"
             ],
             "name" : "Main",
             "program_sources" : [
                "5"
             ],
             "dimensions" : {
                "width" : 1920,
                "height" : 1080
             }
          },
        }
     }

#### *GET* `/api/sources`

#### *GET* `/api/sources/:id`

#### *GET* `/api/screens`

#### *GET* `/api/screens/`

Includes the detailed information for each screen. This is `O(N)` RPCs on the number of screen destinations.

#### *GET* `/api/screens/:id`

#### *GET* `/api/auxes`

#### *GET* `/api/auxes/:id`

#### *GET* `/api/presets`

#### *GET* `/api/presets/`

Includes the detailed information for each preset. This is `O(N)` RPCs on the number of presets.

#### *GET* `/api/presets/:id`

### Events

TODO: examples

#### `ws://.../events`

TODO: support JSON encoding `Client.System`

    {
        "line": "...",
    }

### Usage
    server [OPTIONS]

    Application Options:
          --http-listen=[HOST]:PORT
          --http-static=PATH

    E2 Discovery:
          --discovery-address=
          --discovery-interface=
          --discovery-interval=

    E2 JSON-RPC:
          --e2-address=HOST
          --e2-jsonrpc-port=PORT
          --e2-xml-port=PORT
          --e2-timeout=

## Client
    
    go get ./cmd/server

Useful for testing the client library:

    $GOPATH/bin/client --e2-address=192.168.0.100 listen

### Support

The client library (`github.com/qmsk/e2/client`) supports:
    * The streaming XML API (read-only)
    * The JSON-RPC API

The discovery library (`github.com/qmsk/e2/discovery`) supports UDP broadcast discovery of connected E2 systems.

### Usage:

    client [OPTIONS] <command>

    E2 JSON-RPC:
          --e2-address=HOST
          --e2-jsonrpc-port=PORT
          --e2-xml-port=PORT
          --e2-timeout=

    Help Options:
      -h, --help                    Show this help message

    Available commands:
      aux-list           List Aux destinations
      discover           Discover available E2 systems
      list-destinations  List destinations
      listen             Listen XML packets
      preset-list        List presets
      preset-show        Show preset destinations
      screen-list        List Screen destinations
      screen-show        Show screen content
      source-list        List sources

## Legacy

Python implementation; supports loading settings from the HTTP config backup, and using the telnet API to load presets and transition.

The web UI broken, TODO to remove the client/server implementation once re-implemented.

    PYTHONPATH=../qmsk-dmx ./opt/bin/python3 ./qmsk-e2-web \
        --e2-host 192.168.0.201 \
        --e2-presets-xml etc/xml/ \
        --e2-presets-db var/e2.db \
        -v --debug-module qmsk.net.e2.presets
