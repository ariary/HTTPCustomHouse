package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strings"
)

//UTILS

var (
	Info = Teal
	Warn = Yellow
	Evil = Red
	Good = Green
	Code = Cyan
)

var (
	Black         = Color("\033[1;30m%s\033[0m")
	Red           = Color("\033[1;31m%s\033[0m")
	Green         = Color("\033[1;32m%s\033[0m")
	Yellow        = Color("\033[1;33m%s\033[0m")
	Purple        = Color("\033[1;34m%s\033[0m")
	Magenta       = Color("\033[1;35m%s\033[0m")
	Teal          = Color("\033[1;36m%s\033[0m")
	White         = Color("\033[1;37m%s\033[0m")
	Cyan          = Color("\033[1;96m%s\033[0m")
	Underlined    = Color("\033[4m%s\033[24m")
	Bold          = Color("\033[1m%s\033[0m")
	Italic        = Color("\033[3m%s\033[0m")
	RedForeground = Color("\033[1;41m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

const usage = `Usage of customOfficer:
  -r, --residues display residues of the request not treated by the custom officer
  -h, --help prints help information 
`

func main() {

	in := bufio.NewReader(os.Stdin)
	//request, err := http.ReadRequest(in)
	reader := bufio.NewReader(in)
	tp := textproto.NewReader(reader)

	// First line: POST /index.html HTTP/1.0 or other
	var s string
	var err error
	if s, err = tp.ReadLine(); err != nil {
		fmt.Println("ReadLine", err)
	}
	fmt.Println(s) //TO DO: check if this a POST request and with HTTP 1.1

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Fatal(err)
	}
	// http.Header and textproto.MIMEHeader are both just a map[string][]string
	httpHeader := http.Header(mimeHeader)
	// print header
	for h, v := range httpHeader {
		fmt.Printf("%s: %s\n", h, v[0]) //TO DO handle where multiple value are found for a specific header
	}

	// Get Body with Transfer-Encoding
	bodyB, _ := io.ReadAll(tp.R)           // get body
	bodyB = append([]byte("\n"), bodyB...) //\n of REQUEST is not counted in body => add it
	sTransferEncoding := httpHeader.Get("Transfer-encoding")
	if sTransferEncoding == "chunked" {
		//read body till 0
		endChunk := strings.Index(string(bodyB), "\n0\n")
		if endChunk == -1 {
			log.Fatal("Failed to retrieve end of chunks in request('\\n0\\n')")
		}
		bodyTE := string(bodyB[:endChunk+3]) //+3: take into account \n0\n as EndChunk return the index of the substring beginning
		fmt.Printf(bodyTE)

		if len(bodyB) >= endChunk+3 {
			bodyEnding := string(bodyB[endChunk+3:])
			fmt.Fprintf(os.Stderr, Purple(bodyEnding))
		}
	} else {
		fmt.Print(string(bodyB))
	}
}
