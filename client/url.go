package client

import(
    "net/url"
)

type URL url.URL

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
