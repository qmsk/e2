package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testXML struct {
	fileGlob  string
	fileDelay time.Duration
}

// Create a new XMLClient, connected to a local TCP server that serves up the globbed .xml files in order, using a 1
func testXMLClient(test testXML) *XMLClient {
	streamFiles, err := filepath.Glob(test.fileGlob)
	if err != nil {
		panic(fmt.Errorf("filepath.Glob: %v\n", err))
	}

	// setup mock server
	tcpListener, err := net.ListenTCP("tcp", nil)
	if err != nil {
		panic(err)
	}

	go func(tcpListener *net.TCPListener) {
		for {
			tcpConn, err := tcpListener.AcceptTCP()
			if err != nil {
				panic(err)
			}

			// write each .xml file into the testConn end of the pipe, for XMLClient to read
			go func(tcpConn *net.TCPConn) {
				defer tcpConn.CloseWrite()

				for _, xmlPath := range streamFiles {
					xmlFile, err := os.Open(xmlPath)
					if err != nil {
						panic(fmt.Errorf("os.Open: %v\n", err))
					}
					defer xmlFile.Close()

					if write, err := io.Copy(tcpConn, xmlFile); err != nil {
						panic(fmt.Errorf("io.Copy: %v\n", err))
					} else {
						log.Printf("Send %v: %d bytes\n", xmlPath, write)
					}

					if test.fileDelay > 0 {
						time.Sleep(test.fileDelay)
					}
				}

				log.Printf("Send: done\n")
			}(tcpConn)

			go func(tcpConn *net.TCPConn) {
				defer tcpConn.Close()

				for {
					buf := make([]byte, 1500)

					if read, err := tcpConn.Read(buf); err != nil {
						log.Printf("Recv error: %v\n", err)
						break
					} else {
						log.Printf("Recv %d bytes: %#v\n", read, string(buf[:read]))
					}
				}

				log.Printf("Recv: done\n")
			}(tcpConn)
		}
	}(tcpListener)

	// connect client to test server
	listenAddr := tcpListener.Addr().(*net.TCPAddr)

	tcpConn, err := net.DialTCP("tcp", nil, listenAddr)
	if err != nil {
		panic(err)
	}

	xmlClient := XMLClient{
		timeout: 10 * time.Second,
		conn:    tcpConn,
	}

	if err := xmlClient.start(); err != nil {
		panic(err)
	}

	return &xmlClient
}

// Create an XMLClient, and return the final System state after reading all the .xml files
func testXMLClientRead(test testXML) System {
	xmlClient := testXMLClient(test)

	var system System

	for {
		if readSystem, err := xmlClient.Read(); err == io.EOF {
			return system
		} else if err != nil {
			panic(err)
		} else {
			system = readSystem
		}
	}
}

func TestXmlRead(t *testing.T) {
	// test with some minor concurrency
	xmlClient := testXMLClient(testXML{
		fileGlob:  "./test-xml/test1-*.xml",
		fileDelay: 10 * time.Millisecond,
	})

	listenChan, err := xmlClient.Listen()
	if err != nil {
		t.Fatalf("xmlClient.Listen: %v", err)
	}

	// read System state updates, and updates this to be the final System state after all cumulative updates
	var listenSystem System

	for listenSystem = range listenChan {
		// read it
		_ = listenSystem.String()
	}

	if err := xmlClient.ListenError(); err != nil {
		t.Fatalf("xmlClient.ListenError: %v\n", err)
	} else {
		t.Logf("End of Listen\n")
	}

	// check resulting system state
	if source0, exists := listenSystem.SrcMgr.SourceCol[0]; !exists {
		t.Errorf("Source #0 does not exist")
	} else {
		t.Logf("Source #0: %#v\n", source0)

		if source0.Name != "PC 1-1" {
			t.Errorf("Source #0 Name: %v", source0.Name)
		}
	}

	if screen0, exists := listenSystem.DestMgr.ScreenDestCol[0]; !exists {
		t.Errorf("Screen #0 does not exist")
	} else {
		t.Logf("Screen #0: %#v\n", screen0)

		if screen0.Name != "ScreenDest1" {
			t.Errorf("Screen #0 Name: %v", screen0.Name)
		}

		if screen0.IsActive != 1 {
			t.Errorf("Screen #0: IsActive=%v", screen0.IsActive)
		}

		if layer0, exists := screen0.LayerCollection[0]; !exists {
			t.Errorf("Layer #0 does not exist")
		} else {
			t.Logf("Layer #0: %#v\n", layer0)

			if layer0.LastSrcIdx != 0 {
				t.Errorf("Layer #0: LastSrcIdx=%v", layer0.LastSrcIdx)
			}

			if layer0.PgmMode != 1 || layer0.PvwMode != 0 {
				t.Errorf("Layer #0: PgmMode=%v PvwMode=%v", layer0.PgmMode, layer0.PvwMode)
			}
		}

		if layer1, exists := screen0.LayerCollection[1]; !exists {
			t.Errorf("Layer #1 does not exist")
		} else {
			t.Logf("Layer #1: %#v\n", layer1)

			if layer1.LastSrcIdx != 0 {
				t.Errorf("Layer #1: LastSrcIdx=%v", layer1.LastSrcIdx)
			}

			if layer1.PgmMode != 0 || layer1.PvwMode != 1 {
				t.Errorf("Layer #0: PgmMode=%v PvwMode=%v", layer1.PgmMode, layer1.PvwMode)
			}
		}
	}

	if preset0, exists := listenSystem.PresetMgr.Preset[0]; !exists {
		t.Errorf("Preset #0 does not exist")
	} else {
		t.Logf("Preset #0: %#v\n", preset0)

		if preset0.Name != "Test 1" {
			t.Errorf("Preset #0 Name: %v", preset0.Name)
		}
	}
}

// Adding new outputs/inputs from initial empty state
func TestXmlAdd(t *testing.T) {
	system := testXMLClientRead(testXML{
		fileGlob: "./test-xml/test2-*.xml",
	})

	// check resulting system state
	if source0, exists := system.SrcMgr.SourceCol[0]; !exists {
		t.Errorf("Source #0 does not exist")
	} else {
		if source0.Name != "ScreenDest1_PGM-1" {
			t.Errorf("Source #0 Name: %v", source0.Name)
		}
		if !(source0.InputCfgIndex == -1 && source0.StillIndex == -1 && source0.DestIndex == 0) {
			t.Errorf("Source 0: InputCfgIndex=%v StillIndex=%v DestIndex=%v", source0.InputCfgIndex, source0.StillIndex, source0.DestIndex)
		}
	}
	if source1, exists := system.SrcMgr.SourceCol[1]; !exists {
		t.Errorf("Source #1 does not exist")
	} else {
		if source1.Name != "Input1-2" {
			t.Errorf("Source #1 Name: %v", source1.Name)
		}
		if !(source1.InputCfgIndex == 0 && source1.StillIndex == -1 && source1.DestIndex == -1) {
			t.Errorf("Source 1: InputCfgIndex=%v StillIndex=%v DestIndex=%v", source1.InputCfgIndex, source1.StillIndex, source1.DestIndex)
		}
	}
	if source2, exists := system.SrcMgr.SourceCol[2]; !exists {

	} else {
		t.Errorf("Source #2: still exists: %#v\n", source2)
	}

	if screen0, exists := system.DestMgr.ScreenDestCol[0]; !exists {
		t.Errorf("Screen #0 does not exist")
	} else {
		if screen0.Name != "ScreenDest1" {
			t.Errorf("Screen #0 Name: %v", screen0.Name)
		}

		if screen0.IsActive != 0 {
			t.Errorf("Screen #0: IsActive=%v", screen0.IsActive)
		}
	}
}

// Test initial system state with UserKeys and Presets, and then build + save a new Preset
func TestXmlPresets(t *testing.T) {
	system := testXMLClientRead(testXML{
		fileGlob: "./test-xml/test3-*.xml",
	})

	if system.PresetMgr.LastRecall != 0 {
		t.Errorf("Preset LastRecall=%d", system.PresetMgr.LastRecall)
	}

	if preset0, exists := system.PresetMgr.Preset[0]; !exists {

	} else {
		t.Errorf("Preset #0 should not exist: %#v", preset0)
	}

	if preset1, exists := system.PresetMgr.Preset[1]; !exists {
		t.Errorf("Preset #1 does not exist")
	} else {
		t.Logf("Preset #1: %#v\n", preset1)

		if preset1.Name != "Test 2" {
			t.Errorf("Preset #0 Name: %v", preset1.Name)
		}
	}
}

// Test initial system state with version 3.1
func TestXmlVersion31(t *testing.T) {
	system := testXMLClientRead(testXML{
		fileGlob: "./test-xml/test4-*.xml",
	})

	if system.PresetMgr.LastRecall != -1 {
		t.Errorf("Preset LastRecall=%d", system.PresetMgr.LastRecall)
	}
}
