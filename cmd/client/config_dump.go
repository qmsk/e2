package main

import (
	"github.com/qmsk/e2/client"
	"log"
	"os"
)

type ConfigDump struct {
	SettingsXML string `long:"settings-xml" value-name:"PATH"`
}

func init() {
	parser.AddCommand("config-dump", "Dump XML config", "", &ConfigDump{})
}

func (cmd *ConfigDump) Execute(args []string) error {
	if system, err := client.LoadSettingsFile(cmd.SettingsXML); err != nil {
		return err
	} else {
		log.Printf("Loaded: %v\n", cmd.SettingsXML)

		system.Print(os.Stdout)
	}

	return nil
}
