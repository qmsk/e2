# qmsk-e2
E2 Client, Tally system, (partial) Web UI

## Supported Devices

This implementation has been tested with the following device software versions:

* E2 version 3.2
* S3 version 3.2

This implementation supports the following device APIs:

* TCP port 9876 XML (read-only, live streaming)
* TCP port 9999 JSON-RPC (read-only for now, includes preset destinations)
* UDP port 40961 discovery

## Tally

Tally implementation for following which inputs are active on destinations.
The tally process connects to any discovered E2 systems, and follows their system state using a read-only
implementation of the XML protocol.
Input sources can be tagged with a `tally=ID` in their *Contact* field.
Each tally ID has *program*, *preview* and *active* status if any input with a matching tally ID is used as the
program or preview source on any screen destination layer, or Aux output.

Multiple E2 systems can be connected, and their input/output state is merged.
Stacked systems are supported, the tally system will only connect to the stack master.
The tally system should be restarted if the stack master changes.

The tally system will indicate an error status if any E2 system is disconnected (stops responding within the `--e2-timeout=`).
The tally system will reconnect to any E2 system once is starts responding to discovery packets again.


## Usage

Build the golang binary:

    go get ./cmd/tally

Tag the relevant inputs in EMTS with `tally=ID` in their Contact details field:

![EMTS Contact field](/docs/tally-emts-contact.png?raw=true)

Run the tally software using a network interface connected to the same network as the E2 device:

    $GOPATH/bin/tally --discovery-interface=eth0

## Configuration

Use `--tally-ignore-dest=REGEXP` to ignore matching destinations for tally state.
For example, if you have a local monitor connected to an Aux output named "Monitor", and you do not want to indicate tally inputs as being active on program if they are viewed on the Aux monitor, use `--tally-ignore-dest=Monitor`.

Multiple separate tally systems can be run, using the `--tally-contact-name=` to select the `tally=ID` name used to configure the Input's tally ID.
For example, using `--tally-contact-name=tally-test` would follow the input with a Contact field containing `tally-test=2` as the #2 tally output.

## Outputs

The tally state can be output on a HTTP REST/WebSocket API, on GPIO pins, or RGB LEDs on the SPI bus.

### Web API

The Web output provides an JSON REST API, a JSON WebSocket API:

    tally --http-listen=:8001

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

## Web UI

The Web output also provides an AngularJS frontend using the JSON API:

    tally --http-listen=:8001 --http-static=./static

The `--http-static` is optional, and is only needed for the UI. Use `bower` to prepare the JS deps:

	cd static && bower install

The Web UI uses this WebSocket stream to display a live-updating tally state:

![Tally Web UI](/docs/tally-web.png?raw=true)

Click on the tally number to open up a large single-tally view that can be zoomed as needed:

![Tally View](/docs/tally-pixel.png?raw=true)

The Web UI also includes a list of discovered sources and their status, including possible connection errors:

![Tally Web UI](/docs/tally-sources.png?raw=true)

## GPIO

![Tally GPIO](/docs/tally-gpio.jpg?raw=true)

Support for Linux RaspberryPI GPIO output using `/sys/class/gpio`. Use the `--gpio` options to configure:

      --gpio-green-pin=GPIO-PIN           GPIO pin for green status LED
      --gpio-red-pin=GPIO-PIN             GPIO pin for red status LED
      --gpio-tally-pin=GPIO-PIN           Pass each tally pin as a separate option
      --gpio                              Enable GPIO output

Example:
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

If running the tally binary as root, it will handle exporting and setup of the `/sys/class/gpio` devices itself.
If running as a user in the gpio group, the permissions on the exported GPIO pins must be configured by udev before running the tally binary.
Use the included `cmd/tally/gpio-export.sh ...` script to pre-export the GPIO pins.

The tally program will exit and drive the GPIO pins low on SIGINT. If the tally program crashes or is killed, the GPIO pins will remain stuck
in their previous state. Use the `cmd/tally/gpio-unexport.sh ...` script to clear the GPIO output.

Example systemd service to pre-export the GPIO pins, and drive them low if the tally program exits:

	[Unit]
	Description=github.com/qmsk/e2 tally
	After=network.target

	[Service]
	User=e2-tally
	ExecStartPre=/opt/qmsk-e2/bin/gpio-export.sh 23 24 21 20 16 12 26 19 13 6
	ExecStart=/opt/qmsk-e2/bin/tally \
		--discovery-interface=eth0 \
		--gpio --gpio-green-pin=23 --gpio-red-pin=24 \
		--gpio-tally-pin=21 --gpio-tally-pin=20 --gpio-tally-pin=16 --gpio-tally-pin=12 --gpio-tally-pin=26 --gpio-tally-pin=19 --gpio-tally-pin=13 --gpio-tally-pin=6 \

	KillSignal=SIGINT
	ExecStopPost=/opt/qmsk-e2/bin/gpio-unexport.sh 23 24 21 20 16 12 26 19 13 6

	[Install]
	WantedBy=multi-user.target

## SPI-LED

![Tally APA102 LEDs](/docs/tally-spiled.jpg?raw=true)

Support for APA102 RGB LEDs connected to the Linux RaspberryPI SPI bus using `/dev/spidev`. Use the `--spiled` options to configure:

      --spiled-channel=N                  /dev/spidev0.N
      --spiled-speed=HZ
      --spiled-protocol=apa102|apa102x    Type of LED
      --spiled-count=COUNT                Number of LEDs
      --spiled-debug                      Dump SPI output
      --spiled-intensity=0-255
      --spiled-refresh=HZ
      --spiled-tally-idle=RRGGBB
      --spiled-tally-preview=RRGGBB
      --spiled-tally-program=RRGGBB
      --spiled-tally-both=RRGGBB
      --spiled-status-idle=RRGGBB
      --spiled-status-ok=RRGGBB
      --spiled-status-warn=RRGGBB
      --spiled-status-error=RRGGBB
      --spiled                            Enable SPI-LED output

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

The tally output LED will start pulsing if the input connector is disconnected or has an invalid video signal.

The tally program will exit and drive the SPI LEDs off on SIGINT. If the tally program crashes or is killed, the SPI bus will remain stuck in its previous state.
Use the included `cmd/tally/spiled-down.sh <count>` script to force the SPI LEDs off.

## Server (Web UI)

Web interface for displaying E2 state and controlling presets.

Follow E2 status, providing a HTTP REST + WebSocket JSON API, and an AngularJS web UI:

    go get ./cmd/server
    
    cd static && bower install
    
	$GOPATH/bin/server --discovery-interface=eth0 --http-listen=:8284 --http-static=./static

The server will connect to the first discovered E2 system.

### Usage

	  server [OPTIONS]

	Web:
		  --http-listen=[HOST]:PORT
		  --http-static=PATH

	E2 Discovery:
		  --discovery-address=
		  --discovery-interface=
		  --discovery-interval=

	E2 Client:
		  --e2-address=HOST
		  --e2-jsonrpc-port=PORT
		  --e2-xml-port=PORT
		  --e2-telnet-port=PORT
		  --e2-timeout=
		  --e2-safe                    Safe mode, only modify preview
		  --e2-readonly                Read state, do not modify anything
		  --e2-debug                   Dump commands

	Help Options:
	  -h, --help                       Show this help message

The server can be run in *LIVE*, *SAFE* or *READ* mode.
Use `--e2-readonly` for *READ* mode to limit the client to only commands that read the system state, and do not perform any modifications to preview or program state.
Use `--e2-safe` for *SAFE* mode to limit the client to commands that modify the preview state only, and do not perform any Recall, Cut or AutoTrans commands to program.
The default *LIVE* mode is to allow use of all commands, including those that modify Program state.

The status bar at the top of the UI indicates the current mode, showing green for *LIVE* mode, yellow for *SAFE* mode and grey for *READ* mode.
The status bar will change to read to indicate *ERROR* mode if the WebSocket connection or any REST API request fails.
If the websocket connection is lost, the status bar will turn dark grey.

### Main view

Shows an overview of Sources and their active Preview/Program destinations.

![Web UI Main view](/docs/server-main.png?raw=true)

### Presets

Shows grouped Presets.

![Web UI Presets view](/docs/server-presets.png?raw=true)

Clicking a Preset recalls it to Preview.

There are three different grouping options available. 
*All* will show all presets in ID order.
*X.Y* groups presets by their sequence number.
*PG* groups presets by the console layout's preset pages.

The display size buttons change the size of the preset buttons, between *Small*, *Normal* and *Large*.

Clicking *Take*  transitions the last selected Preset directly to Program.
Enabling *Auto Take* will recall each selected Preset directly to Program. The colors of the Preset buttons change to red to indicate this state.

Clicking *Cut* or *Auto Trans* transitions the current System Preview state to Program.
*WARNING:* This may not necessarily be the same as the last recalled preset, in case other concurrent users are also modifying preview
or the set of active destinations!

The last recalled preset is shown with a border.
Externally recalled presets are colored in white.
Recalling a preset will color the border green. The border is initially dashed until the E2 updates the result of the recall command.
*Take* will change the selected Preset's border to red.
*XXX*: *Cut* and *Auto Trans* do not change the selected Preset's display state.

### System

Raw System state, represented as a collapsible JSON object, live-updated from the `/events` WebSocket:

![Web UI System view](/docs/server-system.png?raw=true)

### API

HTTP REST JSON.

#### *GET* `/api/`

Returns server state:

```
{
    "System": { ... }
}
```

This uses the same JSON format as the WebSocket API.

See the [server.json example file](docs/server.json).

#### *GET* `/api/presets`

List of presets.

#### *POST* `/api/presets` `{ ID: 0, Live: false, Cut: false, AutoTrans: -1 }`

Activate presets on preview/program, and optionally cut/autotrans from preview to program.

Parameter   | Type   | Default       | Meaning
------------|--------|---------------|--------------
`ID`        | `int`  | -1            | Recall preset
`Live`      | `bool` | false         | Recall preset directly to program
`Cut`       | `bool` | false         | Cut preview to program (on active destinations)
`AutoTrans` | `int`  | -1            | Auto transition preview to program, using given number of frames, or 0 for default

#### *GET* `/api/presets/`

List of presets with detailed information about preset destinations.

This is `O(N)` JSON-RPC calls on the number of presets.

#### *GET* `/api/presets/:id`

Detailed information about given preset.

This is a single JSON-RPC call.

### WebSocket `/events`

Live streaming of updated E2 system state on events. Sends the current system state on initial connect.

Uses the same JSON format as `/api/`.

```
{
    "System": { ... }
}
```

See the [server.json example file](docs/server.json).

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
	  config-dump        Dump XML config
	  discover           Discover available E2 systems
	  list-destinations  List destinations
	  listen             Listen XML packets
	  preset-list        List presets
	  preset-show        Show preset destinations
	  screen-list        List Screen destinations
	  screen-show        Show screen content
	  source-list        List sources
