// Tally output to a nixie module from dfrobot:
// http://www.dfrobot.com/index.php?route=product/product&product_id=738
//
// The nixie is controlled by sending 2 bytes via SPI
// This module also communicates with a hetec KVM

package nixie

import (
	"fmt"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/qmsk/e2/hetec-dcp"
	"github.com/qmsk/e2/tally"
	"log"
	"os"
	"sync"
)

// Different numbers the module can display
var nixie_numbers = []uint16{
	0x0008, // 0
	0x1000, // 1
	0x0800, // 2
	0x0400, // 3
	0x0200, // 4
	0x0100, // 5
	0x0080, // 6
	0x0040, // 7
	0x0020, // 8
	0x0010} // 9

const nixie_no_number = uint16(0x0000)   // Number off
const nixie_number_mask = uint16(0x1ff8) // Bitmask for all number bits
const nixie_color_mask = (0xe000)        // Bitmask for the background led colors

// Different colors the background rgb led can display
const nixie_red = uint16(0xa000)
const nixie_blue = uint16(0x6000)
const nixie_green = uint16(0xc000)
const nixie_magenta = uint16(0x2000)
const nixie_cyan = uint16(0x4000)
const nixie_yellow = uint16(0x8000)
const nixie_white = uint16(0x0000)
const nixie_led_off = uint16(0xe000)

type Options struct {
	LivePin    string      `long:"gpio-live-pin"`
	KvmOptions dcp.Options `group:"Hetec DCP Serial client"`
}

func (options Options) Make() (*Nixie, error) {
	var nixie = Nixie{
		options: options,
	}

	if err := nixie.init(options); err != nil {
		return nil, err
	}

	return &nixie, nil
}

type Nixie struct {
	options Options

	//Livepin is HIGH if the kvm channel is live
	livePin embd.DigitalPin

	spiBus     embd.SPIBus
	kvmConsole int
	kvmTallies [4]bool

	kvmChan   chan int
	tallyChan chan tally.State
	closeChan chan bool
	waitGroup sync.WaitGroup
}

func (nixie *Nixie) init(options Options) error {
	fmt.Printf("Init GPIO\n")

	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	if err := embd.InitSPI(); err != nil {
		panic(err)
	}

	fmt.Printf("Intialize SPI\n")
	nixie.spiBus = embd.NewSPIBus(embd.SPIMode0, 0, 50, 8, 100)

	nixie.Send(nixie_no_number | nixie_white)
	nixie.Send(nixie_no_number | nixie_white)
	nixie.Send(nixie_no_number | nixie_white)

	if options.LivePin == "" {

	} else if pin, err := embd.NewDigitalPin(options.LivePin); err != nil {
		return fmt.Errorf("embd.NewDigitalPin %v: %v", options.LivePin, err)

		// Writing as "out" defaults to initializing the value as low.
	} else if err := pin.SetDirection(embd.Out); err != nil {
		return fmt.Errorf("pin.SetDirection %v: %v", options.LivePin, err)
	} else {
		nixie.livePin = pin
	}

	nixie.kvmConsole = 5
	nixie.kvmChan = make(chan int)

	nixie.closeChan = make(chan bool)

	return nil
}

func (nixie *Nixie) RegisterTally(t *tally.Tally) {
	nixie.tallyChan = make(chan tally.State)
	nixie.waitGroup.Add(1)

	go nixie.run()

	t.Register(nixie.tallyChan)
}

func (nixie *Nixie) close() {
	defer nixie.waitGroup.Done()

	log.Printf("Nixie: Close pins and SPI bus..")

	if nixie.livePin != nil {
		nixie.livePin.Close()
	}

	// Turn off the nixie tube and release the spi bus
	nixie.Send(nixie_no_number | nixie_led_off)
	nixie.spiBus.Close()
	embd.CloseSPI()
}

func (nixie *Nixie) updateTally(state tally.State) {
	log.Printf("Nixie: Update tally State:")
	fmt.Printf("\tKVM tallies: ")
	for id := tally.ID(1); id < 5; id++ {
		var pinState = false

		if status, exists := state.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			if status.Status.Program {
				pinState = true
			}
		}
		fmt.Printf("%d: %t ", id, pinState)
		nixie.kvmTallies[id-1] = bool(pinState)
	}
	fmt.Printf("\n")
}

func (nixie *Nixie) run() {
	defer nixie.close()

	go nixie.listenKvm()

	// Initialize the nixie tube
	log.Printf("Initialize nixie tube")
	nixie.Send(nixie_numbers[0] | nixie_yellow)

	log.Printf("Entering message loop")
	for {
		select {
		case nixie.kvmConsole = <-nixie.kvmChan:
			log.Printf("KVM console: %d", nixie.kvmConsole)
		case state, open := <-nixie.tallyChan:
			if !open {
				log.Printf("Tally channel closed")
				return
			}
			nixie.updateTally(state)
		case _ = <-nixie.closeChan:
			log.Printf("Nixie: Done")
			return
		}
		if nixie.kvmConsole < 4 && nixie.kvmConsole >= 0 {
			fmt.Printf("\tKVM console: %d\n", nixie.kvmConsole)
			color := nixie_red
			if nixie.kvmTallies[nixie.kvmConsole] {
				log.Println("KVM console is LIVE")
				nixie.livePin.Write(1)
			} else {
				color = nixie_blue
				log.Println("KVM console is safe")
				nixie.livePin.Write(0)
			}
			nixie.Send(nixie_numbers[nixie.kvmConsole+1] | color)
		} else {
			fmt.Printf("\tKVM console: UNKNOWN\n")
			// Invalid state, just re-init the nixie
			nixie.Send(nixie_numbers[0] | nixie_yellow)
		}
	}
}

func (nixie *Nixie) listenKvm() {
	defer close(nixie.kvmChan)

	if client, err := nixie.options.KvmOptions.Client(); err != nil {
		return err
	} else {
		for {
			if dcpDevice, err := client.Read(); err != nil {
				log.Fatalf("dcp:Client.Read: %v\n", err)
			} else {
				dcpDevice.Print(os.Stdout)
				nixie.kvmChan <- dcpDevice.Mode.Console.Channel
			}
		}
	}
}

func (nixie *Nixie) Send(data uint16) {
	data_buf := []uint8{uint8(data >> 8), uint8(data)}
	fmt.Printf("Sending: %08b%08b\n", data_buf[0], data_buf[1])
	if err := nixie.spiBus.TransferAndReceiveData(data_buf); err != nil {
		panic(err)
	}
}

// Close and Wait..
func (nixie *Nixie) Close() {
	log.Printf("GPIO: Close..")

	nixie.closeChan <- true

	if nixie.tallyChan != nil {
		close(nixie.tallyChan)
	}

	nixie.waitGroup.Wait()
}
