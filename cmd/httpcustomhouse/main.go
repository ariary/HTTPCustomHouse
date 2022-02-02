package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"HTTPCustomHouse/pkg/parser"
	"HTTPCustomHouse/pkg/utils"
)

const usage = `Usage of httpcustomhouse:
  -r, --residues display residues of the request not treated by the custom officer
  -cl, --Content-Length stop request treatment according to Content-Length header value
  -te, --Transfer-Encoding stop request treatment according to chunked encoding
  -h, --help prints help information 
`

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

	if isTE { //TE custom house

		// Get Body with Transfer-Encoding
		sTransferEncoding := httpHeader.Get("Transfer-encoding")
		if sTransferEncoding == "chunked" {
			// TODO: implement real chunked encoding
			// read body till 0
			endChunk := strings.Index(string(bodyB), "0\r\n\r\n") //0\r\n\r\n
			if endChunk == -1 {
				log.Fatal("Failed to retrieve end of chunks in request('0\\r\\n\\r\\n')")
			}
			bodyTE := string(bodyB[:endChunk+5]) //+5: take into account 0\r\n\r\n as EndChunk return the index of the substring beginning
			// 0\r\n\r\n + 1 = 5  .. Why 1 ?
			fmt.Printf(bodyTE)

			if residue && len(bodyB) >= endChunk+5 {
				bodyEnding := string(bodyB[endChunk+5:])
				fmt.Fprintf(os.Stderr, utils.Purple(bodyEnding))
			}
		} else {
			fmt.Print(string(bodyB))
		}
	} else { //CL custom house
		// Get Content-Length value
		sContentLength := httpHeader.Get("Content-Length")
		if sContentLength == "" {
			fmt.Fprintf(os.Stderr, "Content-Length not found")
		}
		contentLength, err := strconv.Atoi(sContentLength)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to convert Content-Length: %s", err)
		}

		// Print request body  as it would be interpreted by server using Content-Length
		bodyCL := string(bodyB[:contentLength+1]) // -1? due to the \n beginning the body form (see above)
		fmt.Printf(bodyCL)

		// Print request residue
		if residue && len(bodyB) >= contentLength+1 {
			bodyResidue := string(bodyB[contentLength+1:])
			fmt.Fprintf(os.Stderr, utils.Purple(bodyResidue))
		}
	}

}
