package config

import "github.com/ariary/HTTPCustomHouse/pkg/request"

type ClientConfig struct {
	Insecure  bool
	Verbose   bool
	Debug     bool
	Follow    bool
	InBrowser bool
	Tls       bool
	AddrPort  string
	Request   request.Request
}
