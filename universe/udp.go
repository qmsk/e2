package universe

import (
  "fmt"
  "net"
  "github.com/qmsk/e2/tally"
  "log"
  "sync"
)

type UDPOptions struct {
    TallyOptions

    Port    string    `long:"universe-udp-port" value-name:"PORT" default:"3050"`
    Addr    string    `long:"universe-udp-addr" value-name:"IP"`
}

func (options UDPOptions) Enabled() bool {
  return (options.Addr != "")
}

func (options UDPOptions) UDPTally() (*UDPTally, error) {
  var udpTally = UDPTally{
    options: options,
  }

  addr := net.JoinHostPort(options.Addr, options.Port)

  if udpAddr, err := net.ResolveUDPAddr("udp", addr); err != nil {
    return nil, fmt.Errorf("ResolveUDPAddr %v: %v", addr, err)
  } else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
    return nil, fmt.Errorf("DialUDP %v: %v", udpAddr, err)
  } else {
    udpTally.udpConn = udpConn
  }

  if tallyConfig, err := options.TallyOptions.TallyConfig(&udpTally); err != nil {
    return nil, err
  } else {
    udpTally.config = tallyConfig
  }

  return &udpTally, nil
}

type UDPTally struct {
  options UDPOptions
  config  *TallyConfig

  udpConn *net.UDPConn

  tallyChan chan tally.State

  closeWG sync.WaitGroup
}

func (udpTally *UDPTally) String() string {
  return udpTally.udpConn.RemoteAddr().String()
}

func (udpTally *UDPTally) Send(msg []byte) error {
  log.Printf("universe:UDPTally %v: send %v", udpTally, string(msg))

  if _, err := udpTally.udpConn.Write(msg); err != nil {
    return err
  }

  return nil
}

func (udpTally *UDPTally) RegisterTally(t *tally.Tally) {
	udpTally.tallyChan = make(chan tally.State)

  udpTally.closeWG.Add(1)
	go udpTally.run()

	t.Register(udpTally.tallyChan)
}

func (udpTally *UDPTally) close() {
  defer udpTally.closeWG.Done()

  if err := udpTally.udpConn.Close(); err != nil {
    log.Printf("universe:UDPTally %v: close: %v", udpTally, err)
  }
}

func (udpTally *UDPTally) run() {
  defer udpTally.close()

  for tallyState := range udpTally.tallyChan {
    if err := udpTally.updateTally(tallyState); err != nil {
      log.Printf("universe:UDPTally %v: updateTally: %v", udpTally, err)
    }
  }

  log.Printf("universe:UDPTally: Done")
}

func (udpTally *UDPTally) updateTally(tallyState tally.State) error {
  log.Printf("universe:UDPTally %v: updateTally...", udpTally)

  return udpTally.config.Execute(tallyState)
}

// Close and Wait..
func (udpTally *UDPTally) Close() {
	log.Printf("universe:UDPTally %v: Close..", udpTally)

	if udpTally.tallyChan != nil {
		close(udpTally.tallyChan)
	}

	udpTally.closeWG.Wait()
}
