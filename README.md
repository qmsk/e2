# qmsk-e2
E2 Client, Tally system, (partial) WebUI

## Supported Devices

This implementation has been tested with the following device software versions:

* E2 version 3.2
* S3 version 3.2

This implementation supports the following device APIs:

* TCP port 9876 XML (read-only, live streaming)
* TCP port 9999 JSON-RPC port (read-only for now, includes preset destinations)
* UDP port 40961 discovery

## Tally

Tally implementation for following which inputs are active on destinations.
The tally process connects to any discovered E2 systems, and follows their system state using a read-only
implementation of the XML protocol.
Input sources can be tagged with a `tally=ID` in their *Contact* field.
Each tally ID has *program*, *preview* and *active* status if any input with a matching tally ID is used as the
program or preview source on any screen destination layer, or Aux output.

## Usage

Build the golang binary:

    go get ./cmd/tally

Tag the relevant inputs in EMTS with `tally=ID` in their Contact details field:

![EMTS Contact field](/docs/tally-emts-contact.png?raw=true)

Run the tally software using a network interface connected to the same network as the E2 device:

    $GOPATH/bin/tally --discovery-interface=eth0

## Outputs

The tally state can be output on a HTTP REST/WebSocket API, on GPIO pins, or RGB LEDs on the SPI bus.

### Web API and UI

The Web output provides an JSON REST API, a JSON WebSocket API, and an AngularJS frontend.

    --http-listen=:8001 --http-static=./static

The `--http-static` is optional, and is only needed for the UI.

Example JSON `http://localhost:8001/api/tally` output:

	{
	   "Inputs" : [
		  {
			 "Name" : "DVI 1",
			 "Source" : "192.168.2.102",
			 "ID" : 2,
			 "Status" : "ok"
		  }
	   ],
	   "Errors" : null,
	   "Tally" : [
		  {
			 "Outputs" : [
				{
				   "Name" : "ScreenDest1",
				   "Active" : true,
				   "Program" : true,
				   "Preview" : true,
				   "Source" : "192.168.2.102"
				},
				{
				   "Preview" : true,
				   "Source" : "192.168.2.102",
				   "Program" : true,
				   "Active" : true,
				   "Name" : "ScreenDest2"
				},
				{
				   "Program" : true,
				   "Source" : "192.168.2.102",
				   "Preview" : true,
				   "Name" : "Mon",
				   "Active" : false
				}
			 ],
			 "Inputs" : [
				{
				   "Source" : "192.168.2.102",
				   "ID" : 2,
				   "Status" : "ok",
				   "Name" : "DVI 1"
				}
			 ],
			 "Errors" : null,
			 "Active" : true,
			 "ID" : 2,
			 "Preview" : true,
			 "Program" : true
		  }
	   ]
	}

The same JSON document is also published on `ws://localhost:8001/events` in the `{"tally": { ... }}` format whenever it is updated.
A HTTP `Origin:` header must be set.

The Web UI uses this WebSocket stream to display a live-updating tally state:

![Tally Web UI](/docs/tally-web.png?raw=true)

The Web UI also includes a list of discovered sources and their status, including possible connection errors:

![Tally Web UI](/docs/tally-sources.png?raw=true)

## GPIO

Support for Linux RaspberryPI GPIO output using `/sys/class/gpio`. Use the `--gpio` options to configure:
    
    $GOPATH/bin/tally ... --gpio --gpio-green-pin=23 --gpio-red-pin=24 --gpio-tally-pin=21 --gpio-tally-pin=20 --gpio-tally-pin=16 --gpio-tally-pin=12 --gpio-tally-pin=26 -gpio-tally-pin=19 --gpio-tally-pin=13 --gpio-tally-pin=6

The `--gpio-green-pin=` and `--gpio-red-pin=` are used for the status of the tally system itself:

| Green     | Red      | Status                          |
|-----------|----------|---------------------------------|
| Off       | Off      | Not running.                    |
| Off/Blink | Off      | Discovering.                    |
| On/Blink  | Off      | Connected with tally inputs.    |
| On/Blink  | Blinking | Partially connected.            |
| Off/Blink | Blinking | Reconnecting.                   |

Each `--gpio-tally-pin=` is used for sequentially numbered tally ID output. Passing eight `--gpio-tally-pin=` options will enable tally output for IDs 1, 2, 3, 4, 5, 6, 7 and 8.
The GPIO pin will be set high if the tally input is output on program, and low otherwise.

## SPI-LED

Support for APA102 RGB LEDs connected to the Linux RaspberryPI SPI bus using `/dev/spidev`. Use the `--spiled` options to configure:

      --spiled-channel=            /dev/spidev0.N
      --spiled-speed=
      --spiled-protocol=           Type of LED
      --spiled-count=              Number of LEDs
      --spiled-debug               Dump SPI output
      --spiled-intensity=
      --spiled-refresh=
      --spiled-tally-idle=
      --spiled-tally-preview=
      --spiled-tally-program=
      --spiled-tally-both=
      --spiled-status-idle=
      --spiled-status-ok=
      --spiled-status-warn=
      --spiled-status-error=
      --spiled                     Enable SPI-LED output

Supported protocols:

* `apa102` standard APA102 LEDs
* `apa102x` variant of APA-102 using 0x00 stop frames.

Example:

	tally ... --spiled --spiled-channel=0 --spiled-speed=100000 --spiled-protocol=apa102x --spiled-count=2 --spiled-tally-idle=000010

The first SPI-LED is used as a status LED, use the `--spiled-status-*` options to configure output colors:

| Option                   | Default Color   | Meaning
|--------------------------|-----------------|-----------
| `--spiled-status-idle=`  | 0000ff (Blue)   | Discovering
| `--spiled-status-ok=`    | 00ff00 (Green)  | Connected
| `--spiled-status-warn=`  | ffff00 (Orange) | Partially connected
| `--spiled-status-error=` | ff0000 (Red)    | Disconnected

The remaining LEDs are used for sequentially numbered tally ID output. Use the `--spiled-tally-*` options to configure output colors:

| Option                    | Default Color     | Meaning
|---------------------------|-------------------|------
|                           | Off               | No tally input configured
| `--spiled-tally-idle=`    | 000010 (Dim Blue) | Not active on any outputs
| `--spiled-tally-preview=` | 00ff00 (Green)    | Preview on active destination
| `--spiled-tally-program=` | ff0000 (Red)      | Program on destination
| `--spiled-tally-both=`    | ff4000 (Orange)   | Program on destination, and Preview on active destination

## Server (Web UI)
    
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
    
    go install ./cmd/client

Useful for testing the client library:

    $GOPATH/bin/client --e2-address=192.168.0.100 listen

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
