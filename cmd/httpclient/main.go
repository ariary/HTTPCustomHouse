package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ariary/HTTPCustomHouse/pkg/client"
	"github.com/ariary/HTTPCustomHouse/pkg/config"
	"github.com/ariary/HTTPCustomHouse/pkg/parser"
	"github.com/ariary/HTTPCustomHouse/pkg/request"
	"github.com/ariary/go-utils/pkg/color"
)

const usage = `Usage of httpclient: httpclient [url]
Make http request from raw request. [url] is required and on the form: [protocol]://[addr]:[port]
  -k, --insecure     insecure HTTPS communication
  -i, --include		 include the HTTP response headers in the output
  -v, --verbose	     display sent request (-d to see special characters)
  -L, --location     follow redirects
  -B, --browser      perform current request in browser (-Bc to add Cookie in further request)
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

	flag.BoolVar(&cfg.Follow, "location", false, "Follow redirections")
	flag.BoolVar(&cfg.Follow, "L", false, "Follow redirections")

	flag.BoolVar(&cfg.InBrowser, "browser", false, "Perform current request in browser")
	flag.BoolVar(&cfg.InBrowser, "B", false, "Perform current request in browser")
	flag.BoolVar(&cfg.InBrowserWithCookie, "Bc", false, "Perform current request in browser, include cookie for other request")

	flag.BoolVar(&cfg.Include, "include", false, "Include the HTTP response headers in the output")
	flag.BoolVar(&cfg.Include, "i", false, "Include the HTTP response headers in the output")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	cfg.Url = flag.Arg(0)

	if cfg.Url == "" {
		fmt.Println("Provide an url (required) on the form: [protocol]://[addr]:[port]")
		os.Exit(1)
	}

	cfg.Tls, cfg.AddrPort = parser.ParseUrl(cfg.Url)

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

	if cfg.InBrowser || cfg.InBrowserWithCookie { // in browser
		client.BrowserMode(cfg)
	} else { // in output
		if cfg.Verbose {
			fmt.Println(color.MagentaForeground("------------------------ SEND:"))
			if cfg.Debug {
				reqDebug := parser.ReplaceSpecialCharacters(rawRequest)
				fmt.Println(reqDebug)
			} else {
				fmt.Println(string(rawRequest)) // raw request ~ request.GetRawRequest(cfg.Request)
			}
			fmt.Println(color.BlueForeground("--------------------- RECEIVE:"))
		}

		respText := client.PerformRequest(cfg)

		response, err := parser.ParseResponse(cfg.Request.Method, cfg.AddrPort, respText)
		if err != nil {
			log.Fatal("Failed parsing response:", err)
		}

		if cfg.Follow {
			if cfg.Verbose {
				fmt.Println(respText)
				fmt.Println("********************* REDIRECT:")
			}
			redirectResponseText, err := client.Redirect(&cfg, response)
			if err != nil {
				fmt.Println(respText)
				log.Fatal(err)
			}
			rawRequest = request.GetRawHTTPRequest(cfg.Request)
			if cfg.Verbose {
				fmt.Println("--------------------- SEND:")
				if cfg.Debug {
					reqDebug := parser.ReplaceSpecialCharacters(rawRequest)
					fmt.Println(reqDebug)
				} else {
					fmt.Println(string(rawRequest)) // raw request ~ request.GetRawRequest(cfg.Request)
				}
				fmt.Println("--------------------- RECEIVE:")
			}
			//from now follow only 1 redirect
			fmt.Println(redirectResponseText)
		} else {
			if cfg.Include {
				fmt.Println(respText)
			} else {
				fmt.Println(string(response.Body))
			}
		}
	}
}
