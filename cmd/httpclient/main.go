package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ariary/HTTPCustomHouse/pkg/config"
	"github.com/ariary/HTTPCustomHouse/pkg/parser"
	"github.com/ariary/HTTPCustomHouse/pkg/request"
	"github.com/ariary/HTTPCustomHouse/pkg/utils"
)

const usage = `Usage of httpclient: httpclient [url]
Make http request from raw request. [url] is required and on the form: [protocol]://[addr]:[port]
  -k, --insecure     insecure HTTPS communication
  -v, --verbose	     display sent request (-d to see special characters)
  -h, --help         prints help information 
`

//REWRITE A GO HTTP CLIENT CAUSE net/http ONE REWRITE HEADERS ETC THAT WE DON'T
//NCAT & OPENSSL S_CLIENT AREN'T SATISFYING EITHER

func main() {
	var cfg config.ClientConfig
	log.SetFlags(log.Lshortfile)

	flag.BoolVar(&cfg.Insecure, "insecure", false, "Insecure HTTPS communication")
	flag.BoolVar(&cfg.Insecure, "k", false, "Insecure HTTPS communication")

	flag.BoolVar(&cfg.Verbose, "verbose", false, "Display sent request")
	flag.BoolVar(&cfg.Verbose, "v", false, "Display sent request")
	flag.BoolVar(&cfg.Debug, "d", false, "Display sent request with special character")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	url := flag.Arg(0)

	if url == "" {
		fmt.Println("Provide an url (required) on the form: [protocol]://[addr]:[port]")
		os.Exit(1)
	}
	cfg.Tls, cfg.AddrPort = parser.ParseUrl(url)

	var err error
	rawRequest, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// parse request
	r := bufio.NewReader(bytes.NewReader(rawRequest))
	cfg.Request, err = parser.ParseRequestWithoutPrint(r)
	if err != nil {
		log.Fatal("Failed parsing request:", err)
	}

	if cfg.Request.Version == "1.1" && cfg.Request.Method == "GET" { //1.1 spec keep-alive the connection for GET requets (POST?)
		cfg.Request.Headers.Add("Connection", "close")
		rawRequest = request.GetRawHTTPRequest(cfg.Request)
	}

	if cfg.Verbose {
		fmt.Println("--------------------- SEND:")
		if cfg.Debug {
			reqDebug := strings.ReplaceAll(string(rawRequest), "\r", utils.Green("\\r"))
			reqDebug = strings.ReplaceAll(reqDebug, "\n", utils.Green("\\n\n"))
			fmt.Println(reqDebug)
		} else {
			fmt.Println(string(rawRequest)) // raw request ~ request.GetRawRequest(cfg.Request)
		}
		fmt.Println("--------------------- RECEIVE:")
	}

	var conn net.Conn
	if cfg.Tls {
		conf := &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		}
		conn, err = tls.Dial("tcp", cfg.AddrPort, conf)
	} else {
		conn, err = net.Dial("tcp", cfg.AddrPort)
	}

	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	n, err := conn.Write(rawRequest)
	if err != nil {
		log.Println(n, err)
		return
	}

	// //print response
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	respText := buf.String()
	//response, err := parser.ParseResponse(cfg.Request.Method, cfg.AddrPort, respText)
	// if err != nil {
	// 	log.Fatal("Failed parsing response:", err)
	// }

	// switch status := response.Status; {
	// case status >= 301 && status <= 303:
	// 	fmt.Println("Follow redirect using get")
	// case status < 301:
	// 	fmt.Println("nothing")
	// case status > 303:
	// 	fmt.Println("nothing")
	// }

	fmt.Println(respText)
}
