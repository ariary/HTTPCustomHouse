package request

import (
	"fmt"
	"net/http"
	"strings"
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

//SetPath: Change the path uri with the one provided for the HTTP Request (modify first line of the raw request)
func (request *Request) SetPath(path string) {
	commandLineSplitted := strings.Split(request.CommandLine, " ") //(0: medthod, 1:path 2: version)
	commandLineSplitted[1] = path
	request.CommandLine = strings.Join(commandLineSplitted, " ")

}

//SetMethod: Change the method for the HTTP Request (modify first line of the raw request + request obect)
func (request *Request) SetMethod(nMethod string) {
	request.Method = nMethod
	commandLineSplitted := strings.Split(request.CommandLine, " ") //(0: medthod, 1:path 2: version)
	commandLineSplitted[0] = nMethod
	request.CommandLine = strings.Join(commandLineSplitted, " ")

}

// From net http package (+ withdraw sanitization)
// AddCookie adds a cookie to the request. Per RFC 6265 section 5.4,
// AddCookie does not attach more than one Cookie header field. That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
// AddCookie only sanitizes c's name and value, and does not sanitize
// a Cookie header already present in the request.
func (r *Request) AddCookie(c *http.Cookie) {
	//s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	s := fmt.Sprintf("%s=%s", c.Name, c.Value)
	if c := r.Headers.Get("Cookie"); c != "" {
		r.Headers.Set("Cookie", c+"; "+s)
	} else {
		r.Headers.Set("Cookie", s)
	}
}
