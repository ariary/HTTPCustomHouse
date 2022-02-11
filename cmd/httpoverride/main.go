package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ariary/HTTPCustomHouse/pkg/parser"
)

const usage = `Usage of httpoverride:
Modify an HTTP request from stdin. timeline of modificatons: override, then delete, finally add
  -A, --add-header        header to add (in form of name:value). Can be used several times
  -H, --header            header to override or delete (in form of name:value to add header OR to remove header: name: or name). Can be used several times
  -cl, --content-length   modify Content-Length header
  -te, --chunked          add chunked encoding header
  --host                  modify Host header
  -h, --help              prints help information 
`

//Http spec: \r\n at each end of line
//HEADER
//\r\n
//BODY

// arrayFlags: flag that can be used multiple times
type arrayFlags []string

func (i *arrayFlags) String() string {

	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	//-H
	var headers arrayFlags
	flag.Var(&headers, "A", "headers to modify")
	flag.Var(&headers, "add-header", "headers to modify")

	//-O
	var headersOverride arrayFlags
	flag.Var(&headersOverride, "H", "headers to override")
	flag.Var(&headers, "header", "headers to overide")

	//-cl
	var contentLength string
	flag.StringVar(&contentLength, "cl", "", "modify Content Length header value")
	flag.StringVar(&contentLength, "content-length", "", "modify Content Length header value")

	//--host
	var host string
	flag.StringVar(&host, "host", "", "modify Host header value")

	//-te
	var chunked bool
	flag.BoolVar(&chunked, "te", false, "add chunked encoding header")
	flag.BoolVar(&chunked, "chunked", false, "add chunked encoding header")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	in := bufio.NewReader(os.Stdin)
	reader := bufio.NewReader(in)

	httpHeader, bodyB, err := parser.ParseRequest(reader)
	if err != nil {
		log.Fatal("Failed parsing request:", err)
	}

	// parse headers from CLi
	newHeaders, overrideHeaders, delHeaders := parseHeadersFromCLI(headers, headersOverride)
	//Shortcuts
	if contentLength != "" {
		overrideHeaders["Content-Length"] = contentLength
	}
	if host != "" {
		overrideHeaders["Host"] = host
	}
	if chunked {
		overrideHeaders["Transfer-Encoding"] = "chunked"
	}

	// override headers
	for header, value := range overrideHeaders {
		delete(httpHeader, header)
		httpHeader[header] = append(httpHeader[header], value)
	}
	// delete headers
	for i := 0; i < len(delHeaders); i++ {
		delete(httpHeader, delHeaders[i])
	}
	//add headers
	for header, value := range newHeaders {
		httpHeader[header] = append(httpHeader[header], value...) //add without overriding already existing value for header entry
	}

	// Print headers
	// always print Host first even if it normally does not have significancy
	// print 1 of them, delete it and continue
	hosts := httpHeader["Host"]
	if len(hosts) != 0 {
		fmt.Printf("Host: %s\r\n", hosts[0])
	}

	if len(hosts) > 1 { //several Host headers
		httpHeader["Host"] = hosts[1:]
	} else { // only one so remove it from headers as it has already been printed
		delete(httpHeader, "Host")
	}

	// print http header
	for h, values := range httpHeader {
		for i := 0; i < len(values); i++ {
			fmt.Printf("%s: %s\r\n", h, values[i])
		}
	}

	// print body ("\r\n" is already in bodyB)
	fmt.Printf(string(bodyB))

}

func parseHeadersFromCLI(headers arrayFlags, overrideHeaders arrayFlags) (nHeaders map[string][]string, oHeaders map[string]string, dHeaders []string) {
	nHeaders = make(map[string][]string)
	oHeaders = make(map[string]string)

	// fill new headers struct
	for i := 0; i < len(headers); i++ {
		flagSanitize := strings.ReplaceAll(headers[i], " ", "") // withdraw useless space
		headerValue := strings.Split(flagSanitize, ":")
		if len(headerValue) == 2 {
			header := headerValue[0]
			value := headerValue[1]
			if value != "" {
				nHeaders[header] = append(nHeaders[header], value)
			} else {
				log.Fatal("Wrong argument for -H, --header flag: [header]:[value], value can't be empty")
			}
		} else {
			log.Fatal("Wrong argument for -H, --header flag: [header]:[value]")
		}
	}

	// fill overriding headers struct and delete headers
	for i := 0; i < len(overrideHeaders); i++ {
		flagSanitize := strings.ReplaceAll(overrideHeaders[i], " ", "") // withdraw useless space
		headerValue := strings.Split(flagSanitize, ":")
		switch len(headerValue) {
		case 2: //[header]:[value]
			header := headerValue[0]
			value := headerValue[1]
			if value != "" {
				oHeaders[header] = value
			} else {
				dHeaders = append(dHeaders, header)
			}
		case 1:
			dHeaders = append(dHeaders, headerValue[0])
		default:
			log.Fatal("Wrong argument for -O, --header-override flag: [header]:[value] or [header]: or [header]")
		}
	}
	return nHeaders, oHeaders, dHeaders
}
