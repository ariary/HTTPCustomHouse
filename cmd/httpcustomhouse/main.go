package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ariary/HTTPCustomHouse/pkg/parser"
	"github.com/ariary/go-utils/pkg/color"
)

const usage = `Usage of httpcustomhouse:
  -r, --residues display residues of the request not treated by the custom officer
  -cl, --Content-Length stop request treatment according to Content-Length header value
  -te, --Transfer-Encoding stop request treatment according to chunked encoding
  -h, --help prints help information 
`

// /!\ request contain \r\n\r\n characters, when editing w/ vscode for example this character are
// automatically deleted. Use echo -ne "0\r\n\r\n" instead

func main() {
	//-r
	var residue bool
	flag.BoolVar(&residue, "residue", false, "display residue of the request not treated by the custom officer")
	flag.BoolVar(&residue, "r", false, "display residue of the request not treated by the custom officer")
	// -cl
	var isCL bool
	flag.BoolVar(&isCL, "Content-Length", false, "stop request treatment according to Content-Length header value (default)")
	flag.BoolVar(&isCL, "cl", false, "stop request treatment according to Content-Length header value")
	//-te
	var isTE bool
	flag.BoolVar(&isTE, "Transfer-Encoding", false, "stop request treatment according to chunked encoding")
	flag.BoolVar(&isTE, "te", false, "stop request treatment according to chunked encoding")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if isCL && isTE {
		fmt.Println("httpcustomhouse: you can't use -cl and -te in the same time. If none is specified -cl is the default")
		os.Exit(1)
	}

	in := bufio.NewReader(os.Stdin)
	//request, err := http.ReadRequest(in)	//we have to rewrite the method by our own as it process CL check and TE also => err
	reader := bufio.NewReader(in)

	httpHeader, bodyB, err := parser.ParseRequest(reader)
	if err != nil {
		log.Fatal("Failed parsing request:", err)
	}

	// Print header
	for h, v := range httpHeader {
		fmt.Printf("%s: %s\n", h, v[0]) //TODO handle where multiple value are found for a specific header
	}

	// /!\ bodyB include \r\n to end headers section
	fmt.Print("\r\n")
	bodyB = bodyB[2:]

	if isTE { //TE custom house

		// Get Body with Transfer-Encoding
		sTransferEncoding := httpHeader.Get("Transfer-encoding")
		if sTransferEncoding == "chunked" {
			bodyTE, residueB := parser.FilterWithChunkEncoding(bodyB)
			fmt.Print(string(bodyTE))
			if residue {
				fmt.Fprintf(os.Stderr, color.Magenta(string(residueB)))
			}
		} else {
			fmt.Print(string(bodyB))
		}
	} else { //CL custom house
		// Get Content-Length value
		sContentLength := httpHeader.Get("Content-Length")
		if sContentLength == "" {
			//fmt.Fprintf(os.Stderr, "Content-Length not found")
			fmt.Print(string(bodyB)) //Print whole request
		} else {
			contentLength, err := strconv.Atoi(sContentLength)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to convert Content-Length: %s", err)
			}

			bodyCL, residueB, difference := parser.FilterWithContentLength(contentLength, bodyB)
			fmt.Print(string(bodyCL))
			if difference > 0 {
				fmt.Fprintln(os.Stderr, color.Yellow("\nMissing ", difference, " bytes in body"))
			} else if residue {
				fmt.Fprintf(os.Stderr, color.Magenta(string(residueB)))
			}

		}

	}

}
