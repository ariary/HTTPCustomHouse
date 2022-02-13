package config

import "github.com/ariary/HTTPCustomHouse/pkg/request"

type ClientConfig struct {
	Insecure bool
	Verbose  bool
	Debug    bool
	Tls      bool
	AddrPort string
	Request  request.Request
}
