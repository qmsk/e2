# qmsk-e2
E2 Client, Tally system. Web UI with Presets.

Pre-built release binaries can be found under the [Releases](https://github.com/qmsk/e2/releases).

Please refer to the [Wiki](https://github.com/qmsk/e2/wiki) for more detailed documentation, including development and feature guides.

You can also try submitting a [GitHub Issue](https://github.com/qmsk/e2/issues/new?labels=question) for support, which may or may not receive an answer.

## Supported Devices

This implementation has been tested with the following device software versions:

* E2 version 3.2
* S3 version 3.2

This implementation supports the following device APIs:

* TCP port 9876 XML (read-only, live streaming)
* TCP port 9999 JSON-RPC (read-mostly, includes preset recall)
* TCP port 9878 "telnet" (write-only, preset recalls and program cut/autotrans)
* UDP port 40961 discovery

## Building

The project consists of a set Go applications, and a Javascript web frontend. Once built, the binary Go applications + Javascript assets can be distributed and executed without needing to install the development tools and instructions listed here.

Release binaries are built using the Docker-based setup under [build](/build)

### Backend

    go get github.com/qmsk/e2/cmd/...

Building the backend code requires [Go version 1.10](https://golang.org/dl/).

The Go binaries can also be cross-compiled for different platforms, such as building Linux ARM binaries on your laptop for use on a RaspberryPI:

    GOOS=linux GOARCH=arm go build -o bin/linux_arm/server -v github.com/qmsk/e2/cmd/server

### Frontend

    cd static && bower install

Building the frontend code requires:

* [NPM](https://www.npmjs.com/)
* [bower](https://bower.io/)

# Tally

Tally implementation for following the state of inputs on program, preview and active destinations.

## Sources

Supports the following input sources:

* E2
* S3

## Drivers

Supports the following output drivers:

* [HTTP REST JSON API](https://github.com/qmsk/e2/wiki/Tally#web-api)
* [Web UI](https://github.com/qmsk/e2/wiki/Tally#web-ui)
* [GPIO](https://github.com/qmsk/e2/wiki/Tally#gpio)
* [SPI RGB LED](https://github.com/qmsk/e2/wiki/Tally#spi-led)
* [Universe UDP](https://github.com/qmsk/e2/wiki/Universe-Tally)

## Usage

Run the tally software using a network interface connected to the same network as the E2 device:

    $GOPATH/bin/tally --discovery-interface=eth0

Tag the relevant inputs in EMTS with `tally=ID` in their Contact details field:

![EMTS Contact field](https://raw.githubusercontent.com/qmsk/e2/master/docs/tally-emts-contact.png)

Referr to the [Wiki](https://github.com/qmsk/e2/wiki/Tally) for further documentation.

# Server

Web API + frontend for following the E2 state and controlling presets.

	server --discovery-interface=eth0 --http-listen=:8284 --http-static=./static

The server will connect to the first discovered E2 system.

![Server Presets UI](https://raw.githubusercontent.com/qmsk/e2/master/docs/server-presets.png)

Using the server Web UI requires the static assets for the web frontend (see [Building](#building)).

Referr to the [Wiki](https://github.com/qmsk/e2/wiki/Server) for further documentation.

# Client

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
