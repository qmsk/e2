package dcp

import (
    "fmt"
    "io"
)

func (device Device) Print(out io.Writer) {
    fmt.Fprintf(out, "Device: type=%v\n", device.Type)
    fmt.Fprintf(out, "Console mode: channel=%d\n", device.Mode.Console.Channel)
    fmt.Fprintf(out, "Video mode: layout=%s channel=%d\n", device.Mode.Video.Layout, device.Mode.Video.Channel)
}
