package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
)

func ParseRequest(reader *bufio.Reader) (httpHeader http.Header, bodyB []byte, err error) {
	tp := textproto.NewReader(reader)

	// First line: POST /index.html HTTP/1.0 or other
	var s string

	if s, err = tp.ReadLine(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(s) //TO DO: check if this a POST request and with HTTP 1.1

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, nil, err
	}
	// http.Header and textproto.MIMEHeader are both just a map[string][]string
	httpHeader = http.Header(mimeHeader)

	//Get body
	bodyB, err = io.ReadAll(tp.R)
	if err != nil {
		return nil, nil, err
	}
	bodyB = append([]byte("\n"), bodyB...)

	return httpHeader, bodyB, err
}
