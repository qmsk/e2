package universe

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const PORT = "3050"

type LineFormat string

const (
	LineFormatCRLF LineFormat = "\r\n"
	LineFormatCR   LineFormat = "\r"
	LineFormatLF   LineFormat = "\n"
)

func (lf *LineFormat) UnmarshalFlag(value string) error {
	switch strings.ToLower(value) {
	case "crlf":
		*lf = LineFormatCRLF
	case "cr":
		*lf = LineFormatCR
	case "lf":
		*lf = LineFormatLF
	default:
		return fmt.Errorf("Invalid LineFormat: %v", value)
	}

	return nil
}

type tallySender interface {
	Send(msg string) error
	Close() error
}

type TallyURL url.URL

func (u *TallyURL) UnmarshalFlag(value string) error {
	if parseURL, err := url.Parse(value); err != nil {
		return err
	} else {
		*u = TallyURL(*parseURL)
	}

	return nil
}

func (url TallyURL) Addr() string {
	if host, port, _ := net.SplitHostPort(url.Host); host != "" && port != "" {
		return net.JoinHostPort(host, port)
	} else {
		return net.JoinHostPort(url.Host, PORT)
	}
}

func (url TallyURL) tallySender(options TallyOptions) (tallySender, error) {
	switch url.Scheme {
	case "udp":
		return makeUDP(options, url)
	default:
		return nil, fmt.Errorf("Invalid Tally sender scheme: %v", url.Scheme)
	}
}
