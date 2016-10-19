package universe

import (
	"bufio"
	"bytes"
	"log"
	"sync"
	"text/template"

	"github.com/qmsk/e2/tally"
)

const TallyTemplate = `
{{ range $id, $state := .Tally }}
<tally{{$id}}-{{ if $state.Status.Program }}high{{ else }}low{{ end }}>
{{ end }}
`

type TallyOptions struct {
	TemplatePath string `long:"universe-tally-template" value-name:"PATH" description:"Custom template file"`

	LineFormat LineFormat `long:"universe-line-format" value-name:"CR|LF|CRLF" default:"CRLF"`
	UDP        []string   `long:"universe-udp" value-name:"HOST[:PORT]" description:"Send UDP commands"`
	TCP        []string   `long:"universe-tcp" value-name:"HOST[:PORT]" description:"Send TCP commands"`
}

func (options TallyOptions) Enabled() bool {
	return len(options.UDP) > 0 || len(options.TCP) > 0
}

func (options TallyOptions) addSender(tallyDriver *TallyDriver, proto string, addr string) error {
	var url = TallyURL{
		Scheme: proto,
		Host:   addr,
	}

	if tallySender, err := url.tallySender(options); err != nil {
		return err
	} else {
		tallyDriver.addSender(tallySender)
	}

	return nil
}

func (options TallyOptions) TallyDriver() (*TallyDriver, error) {
	var tallyDriver = TallyDriver{}

	if options.TemplatePath == "" {
		if template, err := template.New("universe-tally").Parse(TallyTemplate); err != nil {
			panic(err)
		} else {
			tallyDriver.template = template
		}
	} else {
		if template, err := template.ParseFiles(options.TemplatePath); err != nil {
			return nil, err
		} else {
			tallyDriver.template = template
		}
	}

	for _, addr := range options.UDP {
		if err := options.addSender(&tallyDriver, "udp", addr); err != nil {
			return nil, err
		}
	}
	for _, addr := range options.TCP {
		if err := options.addSender(&tallyDriver, "tcp", addr); err != nil {
			return nil, err
		}
	}

	return &tallyDriver, nil
}

// Configurable tally status output
//
// Each line is sent as a separate message
type TallyDriver struct {
	template *template.Template

	tallyChan chan tally.State
	runWG     sync.WaitGroup

	senders []tallySender
}

func (tallyDriver *TallyDriver) addSender(tallySender tallySender) {
	tallyDriver.senders = append(tallyDriver.senders, tallySender)
}

func (tallyDriver *TallyDriver) Send(msg string) {
	for _, sender := range tallyDriver.senders {
		if err := sender.Send(msg); err != nil {
			log.Printf("universe:Tally: Send %v: %v", sender, err)
		}
	}
}

func (tallyDriver *TallyDriver) RegisterTally(t *tally.Tally) {
	tallyDriver.tallyChan = make(chan tally.State)

	tallyDriver.runWG.Add(1)
	go tallyDriver.run()

	t.Register(tallyDriver.tallyChan)
}

func (tallyDriver *TallyDriver) close() {
	defer tallyDriver.runWG.Done()

	for _, sender := range tallyDriver.senders {
		if err := sender.Close(); err != nil {
			log.Printf("universe:TallyDriver: Close %v: %v", sender, err)
		}
	}
}

func (tallyDriver *TallyDriver) run() {
	defer tallyDriver.close()

	for tallyState := range tallyDriver.tallyChan {
		if err := tallyDriver.updateTally(tallyState); err != nil {
			log.Printf("universe:TallyDriver %v: updateTally: %v", tallyDriver, err)
		}
	}

	log.Printf("universe:TallyDriver: Done")
}

func (tallyDriver *TallyDriver) updateTally(tallyState tally.State) error {
	var buffer bytes.Buffer

	log.Printf("universe:TallyDriver: updateTally...")

	if err := tallyDriver.template.Execute(&buffer, tallyState); err != nil {
		return err
	}

	// send
	var scanner = bufio.NewScanner(&buffer)

	for scanner.Scan() {
		msg := scanner.Text()

		if len(msg) == 0 {
			continue
		}

		tallyDriver.Send(msg)
	}

	return scanner.Err()
}

// Close and Wait..
func (tallyDriver *TallyDriver) Close() {
	log.Printf("universe:TallyDriver %v: Close..", tallyDriver)

	if tallyDriver.tallyChan != nil {
		close(tallyDriver.tallyChan)
	}

	tallyDriver.runWG.Wait()
}
