package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"github.com/ariary/HTTPCustomHouse/pkg/request"
	"github.com/ariary/HTTPCustomHouse/pkg/response"
	"github.com/ariary/HTTPCustomHouse/pkg/utils"
)

//Parse a request to retrieve headers and body
func ParseRequest(reader *bufio.Reader) (httpHeader http.Header, bodyB []byte, err error) {
	tp := textproto.NewReader(reader)

	// First line: POST /index.html HTTP/1.0 or other
	var s string

	if s, err = tp.ReadLine(); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(s) //TO DO: check if this a POST request and with HTTP 1.1
	fmt.Printf("%s\r\n", s) //TO DO: check if this a POST request and with HTTP 1.1

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return nil, nil, err
	}
	// http.Header and textproto.MIMEHeader are both just a map[string][]string
	httpHeader = http.Header(mimeHeader)

	//Get body
	bodyB, err = io.ReadAll(tp.R)
	if err != nil {
		return nil, nil, err
	}
	bodyB = append([]byte("\r\n"), bodyB...)

	return httpHeader, bodyB, err
}

//Parse a request to retrieve headers and body but do not print any information
func ParseRequestWithoutPrint(reader *bufio.Reader) (request request.Request, err error) {
	tp := textproto.NewReader(reader)

	// First line: POST /index.html HTTP/1.0 or other
	var s string

	if s, err = tp.ReadLine(); err != nil {
		log.Fatal(err)
	}
	request.CommandLine = s

	request.Method = strings.Split(s, " ")[0] //if error => Wrong raw packet
	tmpSlash := strings.Split(s, "/")
	if len(tmpSlash) < 3 {
		return request, errors.New("first line of request seems incorrect")
	} else {
		request.Version = tmpSlash[2]
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return request, err
	}
	// http.Header and textproto.MIMEHeader are both just a map[string][]string
	request.Headers = http.Header(mimeHeader)

	//Get body
	request.Body, err = io.ReadAll(tp.R)
	if err != nil {
		return request, err
	}
	//bodyB = append([]byte("\n"), bodyB...)

	return request, err
}

//Parse the body according to chunk encoding
func FilterWithChunkEncoding(body []byte, residue bool) {
	// TODO: implement real chunked encoding
	// read body till 0
	endChunk := strings.Index(string(body), "0\r\n\r\n")
	if endChunk == -1 {
		log.Fatal("Failed to retrieve end of chunks in request('0\\r\\n\\r\\n')")
	}
	bodyTE := string(body[:endChunk+5]) //+5: take into account 0\r\n\r\n as EndChunk return the index of the substring beginning
	// 0\r\n\r\n + 1 = 5  .. Why 1 ?
	fmt.Printf(bodyTE)

	if residue && len(body) >= endChunk+5 {
		bodyEnding := string(body[endChunk+5:])
		fmt.Fprintf(os.Stderr, utils.Purple(bodyEnding))
	}
}

//Parse the body according to Content-Length Header
func FilterWithContentLength(contentLength int, body []byte, residue bool) {

	// 3 cases: CL = body length, CL > body length, CL < body length
	difference := contentLength - len(body)
	switch {
	case difference > 0: // print body + nb of bytes missing
		fmt.Printf(string(body))
		fmt.Fprintln(os.Stderr, utils.Yellow("\nMissing ", difference, " bytes in body"))
	case difference <= 0: //print body + extra body payload (if there is)
		// Print request body  as it would be interpreted by server using Content-Length
		bodyCL := string(body[:contentLength+1]) // -1? due to the \n beginning the body form (see above)
		fmt.Printf(bodyCL)

		// Print request residue
		if residue && len(body) >= contentLength+1 {
			bodyResidue := string(body[contentLength+1:])
			fmt.Fprintf(os.Stderr, utils.Purple(bodyResidue))
		}
	}
}

//ParseURl: parse an url ([protocol]://[addr]:[port]) to retreive protocol and address (address contains port)
func ParseUrl(url string) (tls bool, addr string) {
	if !strings.HasPrefix(url, "http") {
		log.Fatal("Bad url argument want: [protocol]://[addr]:[port]")
	}
	tmp := strings.Split(url, ":")
	if len(tmp) < 2 {
		log.Fatal("Bad url argument want: [protocol]://[addr]:[port]")
	}
	protocol := tmp[0]
	switch protocol {
	case "https":
		tls = true
	default:
		tls = false
	}

	switch len(tmp) {
	case 2: //port is not provided
		url := strings.Split(tmp[1], "/")[2]
		var port string
		if tls {
			port = "443"
		} else {
			port = "80"
		}
		addr = url + ":" + port
	case 3: //port is provided
		url := strings.Split(tmp[1], "/")[2]
		addr = url + ":" + tmp[2] //url:port
	default:
		log.Fatal("Bad url argument want: [protocol]://[addr]:[port]")
	}
	return tls, addr
}

//ParseResponse: parse an HTTP response to retrieve the status line, the header field and the body
func ParseResponse(reqMethod string, url string, resp string) (response response.Response, err error) {
	r := bufio.NewReader(strings.NewReader(resp))
	req, err := http.NewRequest(reqMethod, url, nil)
	if err != nil {
		return response, err
	}

	httpResp, err := http.ReadResponse(r, req)
	if err != nil {
		return response, err
	}

	defer httpResp.Body.Close()
	response.Status = httpResp.StatusCode
	response.Headers = httpResp.Header
	response.Body, err = io.ReadAll(httpResp.Body)
	response.Cookies = httpResp.Cookies()

	return response, err
}
