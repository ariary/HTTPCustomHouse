package request

import (
	"net/http"
)

type Request struct {
	CommandLine string //GET / HTTTP/1.1
	Method      string
	Version     string
	Headers     http.Header
	Body        []byte
}

//Reconstruct the raw request content from a Request struct
func GetRawHTTPRequest(req Request) (rawRequest []byte) {
	// command line
	fullCommandLine := append([]byte(req.CommandLine), []byte("\r\n")...)
	rawRequest = append(rawRequest, fullCommandLine...)

	// headers
	for header, value := range req.Headers {
		for i := 0; i < len(value); i++ {
			headerLine := header + ": " + value[i] + "\r\n"
			rawRequest = append(rawRequest, []byte(headerLine)...)
		}
	}

	// body
	body := append([]byte("\r\n"), req.Body...)
	rawRequest = append(rawRequest, body...)

	return rawRequest
}
