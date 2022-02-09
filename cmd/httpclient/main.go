package main

import (
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
)

const usage = `Usage of httpclient: httpclient [url]
Make http request from raw request. [url] is required and on the form: [protocol]://[addr]:[port]
  -k, --insecure     insecure HTTPS communication
  -h, --help         prints help information 
`

//REWRITE A GO HTTP CLIENT CAUSE net/http ONE REWRITE HEADERS ETC THAT WE DON'T
//NCAT & OPENSSL S_CLIENT AREN'T SATISFYING EITHER
func main() {
	log.SetFlags(log.Lshortfile)

	var insecure bool
	flag.BoolVar(&insecure, "insecure", false, "Insecure HTTPS communication")
	flag.BoolVar(&insecure, "k", false, "Insecure HTTPS communication")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	url := flag.Arg(0)

	if url == "" {
		fmt.Println("Provide an url (required) on the form: [protocol]://[addr]:[port]")
		os.Exit(1)
	}
	encr, addrPort := parseUrl(url)

	req, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var conn net.Conn
	if encr {
		conf := &tls.Config{
			InsecureSkipVerify: insecure,
		}
		conn, err = tls.Dial("tcp", addrPort, conf)
	} else {
		conn, err = net.Dial("tcp", addrPort)
	}

	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	n, err := conn.Write(req)
	if err != nil {
		log.Println(n, err)
		return
	}

	//print response
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	fmt.Println(buf.String())
}

func parseUrl(url string) (tls bool, addr string) {
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
