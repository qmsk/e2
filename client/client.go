package client

import (
	"time"
)

type Options struct {
	Address  string        `long:"e2-address" value-name:"HOST"`
	JSONPort string        `long:"e2-jsonrpc-port" value-name:"PORT" default:"9999"`
	XMLPort  string        `long:"e2-xml-port" value-name:"PORT" default:"9876"`
	TCPPort  string		   `long:"e2-telnet-port" value-name:"PORT" default:"9878"`
	Timeout  time.Duration `long:"e2-timeout" default:"10s"`
	Safe	 bool          `long:"e2-safe" description:"Safe mode, only modify preview"`
	ReadOnly bool          `long:"e2-readonly" description:"Read state, do not modify anything"`
	Debug	 bool		   `long:"e2-debug" description:"Dump commands"`

	ReadKeepalive bool // return keepalive updates from XMLClient.Read()
}

// Returns something sufficient to identify matching Options for the same System
func (options Options) String() string {
	return options.Address
}

type API interface {
	AutoTrans() error
	AutoTransFrames(frames int) error
	Cut() error
	PresetSave(preset Preset) error
	PresetRecall(preset Preset) error
	PresetAutoTrans(preset Preset) error
}
