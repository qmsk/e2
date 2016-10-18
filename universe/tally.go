package universe

import (
  "bufio"
  "bytes"
  "github.com/qmsk/e2/tally"
  "text/template"
)

const TallyTemplate = `
{{ range $id, $state := .Tally }}
<tally{{$id}}-{{ if $state.Status.Program }}high{{ else }}low{{ end }}>
{{ end }}
`

type TallyOptions struct {
  TemplatePath  string `long:"universe-tally-template" value-name:"PATH"`
}

func (options TallyOptions) TallyConfig(sender TallySender) (*TallyConfig, error) {
  var tallyConfig = TallyConfig{
    sender: sender,
  }

  if options.TemplatePath == "" {
    if template, err := template.New("universe-tally").Parse(TallyTemplate); err != nil {
      panic(err)
    } else {
      tallyConfig.template = template
    }
  } else {
    if template, err := template.ParseFiles(options.TemplatePath); err != nil {
      return nil, err
    } else {
      tallyConfig.template = template
    }
  }

  return &tallyConfig, nil
}

type TallySender interface {
  Send(msg []byte) error
}

// Configurable tally status output
//
// Each line is sent as a separate message
type TallyConfig struct {
  sender    TallySender
  template  *template.Template
}

func (tallyConfig *TallyConfig) Execute(tallyState tally.State) error {
  var buffer bytes.Buffer

  if err := tallyConfig.template.Execute(&buffer, tallyState); err != nil {
    return err
  }

  // send
  var scanner = bufio.NewScanner(&buffer)

  for scanner.Scan() {
    msg := scanner.Bytes()

    if len(msg) == 0 {
      continue
    }

    tallyConfig.sender.Send(msg)
  }

  return scanner.Err()
}
