package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ariary/HTTPCustomHouse/pkg/parser"
)

const usage = `Usage of httpoverride:
  -H, --header            headers to modify
  -v, --value             header value
  -d, --delete            delete header
  -A, --add               add header even if it already exists
  -cl, --content-length   modify Content-Length header
  -te, --chunked          add chunked encoding header
  --host                  modify Host header
  -h, --help    prints help information 
`

// /!\ request contain \r\n\r\n characters, when editing w/ vscode for example this character are
// automatically deleted. Use echo -ne "0\r\n\r\n" instead

//Http spec: \r\n at each end of line
//HEADER
//\r\n
//BODY

func main() {
	//-H
	var header string
	flag.StringVar(&header, "header", "Content-Length", "headers to modify")
	flag.StringVar(&header, "H", "Content-Length", "headers to modify")
	//-v
	var value string
	flag.StringVar(&value, "value", "", "header value")
	flag.StringVar(&value, "v", "", "header value")
	//-d
	var del bool
	flag.BoolVar(&del, "d", false, "delete header")
	flag.BoolVar(&del, "delete", false, "delete header")
	//-A
	var add bool
	flag.BoolVar(&add, "A", false, "add header even if it already exists")
	flag.BoolVar(&add, "add", false, "add header even if it already exists")

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

	//Shortcuts
	if contentLength != "" {
		header = "Content-Length"
		value = contentLength
	}

	if host != "" {
		header = "Host"
		value = host
	}

	if chunked {
		header = "Transfer-Encoding"
		value = "chunked"
	}

	if value == "" && !del {
		fmt.Fprintf(os.Stderr, "Please define a value for header with -v flag")
		os.Exit(1)
	}

	if del && (value != "" || add) {
		fmt.Fprintf(os.Stderr, "-d flag can't be used with -A or -v")
		os.Exit(1)
	}

	in := bufio.NewReader(os.Stdin)
	reader := bufio.NewReader(in)

	httpHeader, bodyB, err := parser.ParseRequest(reader)
	if err != nil {
		log.Fatal("Failed parsing request:", err)
	}

	// Modify Header
	if add {
		httpHeader[header] = append(httpHeader[header], value)
	} else if del {
		delete(httpHeader, header)
	} else {
		httpHeader[header] = []string{value}
	}
	// Print header
	// always print Host first
	//print 1 of them, delete it and continue
	hosts := httpHeader["Host"]
	if len(hosts) != 0 {
		fmt.Printf("Host: %s\r\n", hosts[0])
	}

	if len(hosts) > 1 { //plenty Host headers
		httpHeader["Host"] = hosts[1:]
	} else { // only one so remove it from headers as it has already been printed
		delete(httpHeader, "Host")
	}

	for h, values := range httpHeader {
		for i := 0; i < len(values); i++ {
			fmt.Printf("%s: %s\r\n", h, values[i])
		}
	}

	//print body ("\r\n" is already in bodyB)
	fmt.Printf(string(bodyB))

}
