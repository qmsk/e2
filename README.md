# qmsk-e2
E2 Client + REST + WebSocket Server + Web UI

## Web UI

### System
Raw System state, represented as a collapsible JSON object, live-updated from the `/events` WebSocket:

![#/system](/docs/web-system.png?raw=true)

## Server

    go get ./cmd/server
    
    cd static && bower install

Follow E2 status, providing a REST + WebSocket API, and a web UI:

    $GOPATH/bin/server --discovery-interface=eth0 --http-listen=:8284 --http-static=./static

### API

TODO: examples

#### *GET* `/api/`

Combines both sources and screens, including cross-correlated program/preview state for both. This is `O(N)` RPCs on the number of screen destinations.

```JSON
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
         "preview_screens" : [
            "0"
         ],
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
         "program_screens" : [
            "0"
         ],
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
```

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

Supports live streaming of updated E2 system state on events.

The same output can be followed using `client listen` and `client listen --json`.

#### `ws://.../events`

```JSON
{
    "system": {
       "OSVersion" : "0.0.0",
       "PresetMgr" : {
          "LastRecall" : -1,
          "ID" : 0,
          "Preset" : {}
       },
       "DestMgr" : {
          "AuxDestCol" : {},
          "ID" : 0,
          "ScreenDestCol" : {
             "0" : {
                "Transition" : [
                   {
                      "AutoTransInProg" : 0,
                      "ArmMode" : 0,
                      "ID" : 0,
                      "TransInProg" : 0,
                      "TransPos" : 0
                   },
                   {
                      "AutoTransInProg" : 0,
                      "ID" : 1,
                      "ArmMode" : 0,
                      "TransInProg" : 0,
                      "TransPos" : 0
                   }
                ],
                "VSize" : 1080,
                "HSize" : 1920,
                "Name" : "ScreenDest1",
                "IsActive" : 0,
                "LayerCollection" : {
                   "0" : {
                      "LastSrcIdx" : 1,
                      "PgmMode" : 1,
                      "PgmZOrder" : 0,
                      "IsActive" : 0,
                      "PvwMode" : 0,
                      "Name" : "Layer1-A",
                      "id" : 0,
                      "LastUserKeyIdx" : -1,
                      "PvwZOrder" : 0
                   },
                   "3" : {
                      "PvwZOrder" : 2,
                      "LastUserKeyIdx" : -1,
                      "PvwMode" : 0,
                      "Name" : "Layer2-B",
                      "IsActive" : 0,
                      "id" : 3,
                      "PgmMode" : 0,
                      "PgmZOrder" : 2,
                      "LastSrcIdx" : -1
                   },
                   "1" : {
                      "PgmZOrder" : 0,
                      "PgmMode" : 0,
                      "LastSrcIdx" : 1,
                      "LastUserKeyIdx" : -1,
                      "PvwZOrder" : 0,
                      "id" : 1,
                      "PvwMode" : 1,
                      "Name" : "Layer1-B",
                      "IsActive" : 0
                   },
                   "2" : {
                      "PvwZOrder" : 2,
                      "LastUserKeyIdx" : -1,
                      "IsActive" : 0,
                      "Name" : "Layer2-A",
                      "PvwMode" : 0,
                      "id" : 2,
                      "PgmMode" : 0,
                      "PgmZOrder" : 2,
                      "LastSrcIdx" : -1
                   }
                },
                "BGLayer" : [
                   {
                      "Name" : "",
                      "BGShowMatte" : 1,
                      "BGColor" : {
                         "Red" : 0,
                         "Blue" : 0,
                         "Green" : 0,
                         "id" : 0
                      },
                      "id" : 0,
                      "LastBGSourceIndex" : -1
                   },
                   {
                      "LastBGSourceIndex" : -1,
                      "Name" : "",
                      "BGShowMatte" : 1,
                      "BGColor" : {
                         "id" : 0,
                         "Blue" : 0,
                         "Green" : 0,
                         "Red" : 0
                      },
                      "id" : 1
                   }
                ],
                "ID" : 0
             }
          }
       },
       "SrcMgr" : {
          "ID" : 0,
          "SourceCol" : {
             "0" : {
                "HSize" : 1920,
                "StillIndex" : -1,
                "id" : 0,
                "SrcType" : 2,
                "Name" : "ScreenDest1_PGM-1",
                "DestIndex" : 0,
                "UserKeyIndex" : -1,
                "VSize" : 1080,
                "InputCfgVideoStatus" : 0,
                "InputCfgIndex" : -1
             },
             "1" : {
                "DestIndex" : -1,
                "VSize" : 1080,
                "UserKeyIndex" : -1,
                "InputCfgVideoStatus" : 0,
                "InputCfgIndex" : 0,
                "HSize" : 1920,
                "id" : 1,
                "StillIndex" : -1,
                "SrcType" : 0,
                "Name" : "Input1-2"
             }
          },
          "InputCfgCol" : {
             "0" : {
                "id" : 0,
                "InputCfgType" : 0,
                "Name" : "Input1",
                "ConfigOwner" : "",
                "ConfigContact" : "",
                "InputCfgVideoStatus" : 4
             }
          }
       },
       "Version" : "0.0.0",
       "Name" : "System1"
    }
}
```

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

The client library (`github.com/qmsk/e2/client`) provides partial support for the following E2 APIs:
* TCP XML (read-only, streaming)
* JSON-RPC (read-only for now, includes preset destinations)

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
