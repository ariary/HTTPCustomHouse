package config

import "github.com/ariary/HTTPCustomHouse/pkg/request"

type ClientConfig struct {
	Insecure            bool
	Verbose             bool
	Debug               bool
	Follow              bool
	InBrowser           bool
	InBrowserWithCookie bool
	Tls                 bool
	Include             bool
	AddrPort            string //addr:port
	Url                 string //protocol://addr:port
	Request             request.Request
}
