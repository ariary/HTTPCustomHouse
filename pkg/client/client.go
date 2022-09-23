package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ariary/HTTPCustomHouse/pkg/config"
	"github.com/ariary/HTTPCustomHouse/pkg/parser"
	"github.com/ariary/HTTPCustomHouse/pkg/request"
	"github.com/ariary/HTTPCustomHouse/pkg/response"
	encryption "github.com/ariary/go-utils/pkg/encrypt"
	"golang.org/x/net/html"
)

//PerformRequest: Perform the HTTP request providing by the client in argument
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
	port := ":8080"
	// generate random endpoint
	endpoint := "/" + encryption.GenerateRandom()

	fmt.Println("Visit http://localhost" + port + endpoint)
	requestHandler := &RequestHandler{Config: cfg}
	http.HandleFunc(endpoint, requestHandler.RequestWebhookHandler)
	if cfg.InBrowserWithCookie {
		// ~> proxify future requests with cookies
		http.HandleFunc("/", requestHandler.ProxyHandler)
	}

	log.Fatal(http.ListenAndServe(port, nil))
}

type RequestHandler struct {
	Config config.ClientConfig
	Client http.Client
}

//perform a requets when reached
func (rh *RequestHandler) RequestWebhookHandler(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "request on "+rh.Config.AddrPort)
	responseText := PerformRequest(rh.Config)
	response, err := parser.ParseResponse(rh.Config.Request.Method, rh.Config.AddrPort, responseText)
	if err != nil {
		log.Fatal("Failed parsing response:", err)
	}
	body := response.Body
	if rh.Config.InBrowser { // perform request, add base tag in response
		modifiedBody, err := ChangeHTMLBase(body, rh.Config.Url)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, string(modifiedBody))
	} else { //rh.Config.InBrowserWithCookie
		//perform request as usual, resource with relative paths will be fetched with cookie
		fmt.Fprintf(w, string(body))
	}

	fmt.Println("[*] request endpoint webhook reached: forwarded to", rh.Config.Url)
}

//Endpoint to redirect request to other endpoint (in rh.Config object)
// use to proxify traffic between browser and target with cookie, headers etc
func (rh *RequestHandler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	endpoint := rh.Config.AddrPort + "/" + r.URL.Path[1:]

	//modify request
	r.Host = rh.Config.AddrPort

	// Since the r.URL will not have all the information set,
	// such as protocol scheme and host, we create a new URL
	r.RequestURI = "" //mandatory
	u, err := url.Parse(rh.Config.Url)
	if err != nil {
		panic(err)
	}
	r.URL = u

	//TODO: perform request with request.Perform
	// perform request
	delete(r.Header, "Accept-Encoding") //TODO: handle Accept-Encoding https://stackoverflow.com/questions/13130341/reading-gzipped-http-response-in-go

	cookies := rh.Config.Request.Headers["Cookie"]
	for i := 0; i < len(cookies); i++ {
		r.Header.Add("Cookie", cookies[i])
	}

	resp, err := rh.Client.Do(r)
	if err != nil {
		fmt.Println("failed proxifying to", endpoint, ":", err)
		return
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Println("failed reading response body:", err)
		return
	}

	fmt.Println("[~>] request proxyfied to:", endpoint)
}

// ChangeHTMLBase: Browse an HTML file to add base tag in head one.
// If base is already defined, return the original HTML
// If change failed, return original HTML
func ChangeHTMLBase(htmlB []byte, baseUrl string) (nHtml []byte, err error) {
	baseContent := "<base href=" + baseUrl + " />"

	tokenizer := html.NewTokenizer(bytes.NewReader(htmlB))

	//var newHead byte
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			err = tokenizer.Err()
			if err != io.EOF {
				return htmlB, err
			} else {
				return nHtml, nil
			}
		case html.TextToken:
			//newHead +=string(tokenizer.Text())
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "head" { //in head -> add base tag
				nHtml = append(nHtml, tokenizer.Raw()...)
				nHtml = append(nHtml, []byte(baseContent)...) //add <base> in <head>
			} else if string(tn) == "base" { //base is already defined, do not need to redefine it
				return htmlB, err
			} else { //neither in head nor base
				nHtml = append(nHtml, tokenizer.Raw()...)
			}
		default:
			nHtml = append(nHtml, tokenizer.Raw()...)
		}
	}
}

func Redirect(cfg *config.ClientConfig, response response.Response) (redirectResponseText string, err error) {

	switch status := response.Status; {
	case status >= 301 && status <= 303:
		switch location := response.Headers.Get("Location"); {
		case location == "":
			err = errors.New("failed to retrieve Location header in 30X response")
			return "", err
		case strings.HasPrefix(location, "http"):
			isEncrypted, addr := parser.ParseUrl(location)
			if isEncrypted {
				cfg.Tls = true
			} else {
				cfg.Tls = false
			}
			cfg.AddrPort = addr
			path := "/" + strings.Join(strings.Split(location, "/")[4:], "/")
			cfg.Request.SetPath(path)
			//Update Host
			cfg.Request.Headers["Host"] = []string{strings.Split(cfg.AddrPort, ":")[0]}
		default:
			cfg.Request.SetPath(location)
		}

		cfg.Request.SetMethod("GET")
		// add cookie if present
		if cookies := response.Cookies; len(cookies) > 0 {
			for i := 0; i < len(cookies); i++ {
				cfg.Request.AddCookie(cookies[i])
			}
		}
		redirectResponseText = PerformRequest(*cfg)
	case status > 303 && status < 400:
		redirectResponseText = PerformRequest(*cfg)
	// case status > 303:
	// 	fmt.Println("nothing")
	default:
		err = errors.New("status code not treated by redirection:" + strconv.Itoa(status))
		return "", err
	}

	return redirectResponseText, err
}
