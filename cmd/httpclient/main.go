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
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ariary/HTTPCustomHouse/pkg/config"
	"github.com/ariary/HTTPCustomHouse/pkg/parser"
	"github.com/ariary/HTTPCustomHouse/pkg/request"
	"github.com/ariary/HTTPCustomHouse/pkg/utils"
)

const usage = `Usage of httpclient: httpclient [url]
Make http request from raw request. [url] is required and on the form: [protocol]://[addr]:[port]
  -k, --insecure     insecure HTTPS communication
  -v, --verbose	     display sent request (-d to see special characters)
  -L, --location     follow redirects
  -B, --browser      perform current request in browser
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

	if cfg.InBrowser { // in browser
		BrowserMode(cfg)
	} else { // in output
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

		respText := PerformRequest(cfg)

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
}

func PerformRequest(cfg config.ClientConfig) (fullResponseText string) {
	var conn net.Conn
	var err error
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

	rawRequest := request.GetRawHTTPRequest(cfg.Request)
	n, err := conn.Write(rawRequest)
	if err != nil {
		log.Println(n, err)
		return
	}

	// //print response
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	fullResponseText = buf.String()
	return fullResponseText
}

//BrowserMode: Enable to perform request in browser
// Set up a local http server with specific endpoint to reach.
// When this endpoint is reached => request is performed and
// request body returned with <base> tag added if not present
//to redirect each link to original URL
func BrowserMode(cfg config.ClientConfig) {
	endpoint := "/" + generateEndpoint()

	fmt.Println("Visit http://localhost:8080" + endpoint)
	requestHandler := &RequestHandler{Config: cfg}
	http.HandleFunc(endpoint, requestHandler.requestWebhookHandler)
	http.HandleFunc("/", requestHandler.proxyHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

type RequestHandler struct {
	Config config.ClientConfig
	Client http.Client
}

//perform a requets when reached
func (rh *RequestHandler) requestWebhookHandler(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "request on "+rh.Config.AddrPort)
	responseText := PerformRequest(rh.Config)
	response, err := parser.ParseResponse(rh.Config.Request.Method, rh.Config.AddrPort, responseText)
	if err != nil {
		log.Fatal("Failed parsing response:", err)
	}
	body := response.Body
	//TODO add <base> tag
	fmt.Fprintf(w, string(body))
	fmt.Println("[*] request endpoint webhook reached: forwarded to", rh.Config.AddrPort)
}

//Endpoint to redirect request to other endpoint (in rh.Config object)
// use to proxify traffic between browser and target with cookie, headers etc
func (rh *RequestHandler) proxyHandler(w http.ResponseWriter, r *http.Request) {
	endpoint := rh.Config.AddrPort + "/" + r.URL.Path[1:]
	//modify request
	r.Host = rh.Config.AddrPort

	// Since the r.URL will not have all the information set,
	// such as protocol scheme and host, we create a new URL
	r.RequestURI = "" //mandatory
	var urlS string
	if rh.Config.Tls {
		urlS = "https://"
	} else {
		urlS = "http://"
	}
	urlS += endpoint
	u, err := url.Parse(urlS)
	if err != nil {
		panic(err)
	}
	r.URL = u

	//TODO: add Cookie header from rh.Confing.Headers
	//TODO: perform request with request.Perform
	// perform request
	resp, err := rh.Client.Do(r)
	if err != nil {
		fmt.Println("failed proxifying to", endpoint, ":", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed reading body from", endpoint, ":", err)
		return
	}
	fmt.Fprintf(w, string(body))
	fmt.Println("[~>] request proxyfied to:", endpoint)
}

//generateEndpoint: generate a "random" string of 6 alphanumeric charcaters
func generateEndpoint() string {
	rand.Seed(time.Now().UnixNano())
	var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")
	b := make([]rune, 6)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
