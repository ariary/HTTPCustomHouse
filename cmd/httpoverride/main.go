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
  -H, --header  headers to modify
  -v, --value   header value
  -d, --delete  delete header
  -A, --add     add header even if it already exists
  -h, --help    prints help information 
`

// /!\ request contain \r\n\r\n characters, when editing w/ vscode for example this character are
// automatically deleted. Use echo -ne "0\r\n\r\n" instead

func main() {
	//-H
	var header string
	flag.StringVar(&header, "header", "Content-Length", "headers to modify")
	flag.StringVar(&header, "H", "Content-Length", "headers to modify")
	//-H
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
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

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
	for h, values := range httpHeader {
		for i := 0; i < len(values); i++ {
			fmt.Printf("%s: %s\n", h, values[i])
		}
	}

	//print body
	fmt.Print(string(bodyB))

}
