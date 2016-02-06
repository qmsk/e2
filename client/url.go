package client

import (
    "net"
    "net/url"
)

const PORT = "9999"

type URL url.URL

func makeURL(ip net.IP) URL {
    return URL{
        Scheme: "http",
        Host:   net.JoinHostPort(ip.String(), PORT),
    }
}

func (u *URL) UnmarshalFlag(value string) error {
    if parseURL, err := url.Parse(value); err != nil {
        return err
    } else {
        *u = (URL)(*parseURL)
    }

    return nil
}

func (u URL) Empty() bool {
    return u.Scheme == ""
}

func (u URL) String() string {
    return (*url.URL)(&u).String()
}
