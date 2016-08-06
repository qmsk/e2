package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

type System struct {
	OSVersion string
	Version   string
	Name      string

	SrcMgr           SrcMgr
	DestMgr          DestMgr
	PresetMgr        PresetMgr
	FrameCollection  FrameCollection
	ConsoleLayoutMgr ConsoleLayoutMgr
}

func (system *System) Reset() {
	log.Printf("client: System.Reset()\n")

	*system = System{}
}

func (system *System) Print(f io.Writer) {
	for _, inputCfg := range system.SrcMgr.InputCfgCol.List() {
		fmt.Fprintf(f, "Input %d: %s\n\tInputCfgType=%d\n\tInputCfgVideoStatus=%v\n\tConfig: Owner=%#v Contact=%#v\n", inputCfg.ID, inputCfg.Name, inputCfg.InputCfgType, inputCfg.InputCfgVideoStatus, inputCfg.ConfigOwner, inputCfg.ConfigContact)
	}

	for _, source := range system.SrcMgr.SourceCol.List() {
		fmt.Fprintf(f, "Source %d: %v\n\tSrcType: %v\n\tInputCfgIndex=%d StillIndex=%d DestIndex=%d\n", source.ID, source.Name, source.SrcType, source.InputCfgIndex, source.StillIndex, source.DestIndex)
	}

	for _, auxDest := range system.DestMgr.AuxDestCol.List() {
		fmt.Fprintf(f, "Aux %d: %v\n\tISActive: %v\n\tPvwLastSrcIndex=%d PgmLastSrcIndex=%d\n", auxDest.ID, auxDest.Name, auxDest.IsActive, auxDest.PvwLastSrcIndex, auxDest.PgmLastSrcIndex)
	}

	for _, screenDest := range system.DestMgr.ScreenDestCol.List() {
		fmt.Fprintf(f, "Screen %d: %v\n\tIsActive: %v\n\tHSize=%v VSize=%v\n", screenDest.ID, screenDest.Name, screenDest.IsActive, screenDest.HSize, screenDest.VSize)

		for _, layer := range screenDest.LayerCollection.List() {
			fmt.Fprintf(f, "\tLayer %d: %v\n\t\tIsActive=%v\n\t\tPgmMode=%v PvwMode=%v\n\t\tLastSrcIdx=%d\n", layer.ID, layer.Name, layer.IsActive, layer.PgmMode, layer.PvwMode, layer.LastSrcIdx)
		}
	}

	fmt.Fprintf(f, "Presets:\n\tLastRecall: %v\n", system.PresetMgr.LastRecall)

	for _, preset := range system.PresetMgr.Preset.List() {
		fmt.Fprintf(f, "Preset %d: %v\n\tLockMode: %v\n\tSno: %v\n", preset.ID, preset.Name, preset.LockMode, preset.Sno)
	}
}

func (system *System) String() string {
	var buf bytes.Buffer

	system.Print(&buf)

	return buf.String()
}
