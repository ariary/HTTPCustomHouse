package main

import (
	"HTTPCustomHouse/pkg/parser"
	"HTTPCustomHouse/pkg/utils"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const usage = `Usage of customOfficer-TE:
  -r, --residues display residues of the request not treated by the custom officer
  -h, --help prints help information 
`

func main() {

	var residue bool
	flag.BoolVar(&residue, "residue", false, "display residue of the request not treated by the custom officer")
	flag.BoolVar(&residue, "r", false, "display residue of the request not treated by the custom officer")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	in := bufio.NewReader(os.Stdin)
	//request, err := http.ReadRequest(in)	//we have to rewrite the method by our own as it process CL check and TE also => err
	reader := bufio.NewReader(in)

	httpHeader, bodyB, err := parser.ParseRequest(reader)
	if err != nil {
		log.Fatal("Failed parsing request:", err)
	}

	// Print header
	for h, v := range httpHeader {
		fmt.Printf("%s: %s\n", h, v[0]) //TO DO handle where multiple value are found for a specific header
	}

	// Get Body with Transfer-Encoding
	sTransferEncoding := httpHeader.Get("Transfer-encoding")
	if sTransferEncoding == "chunked" {
		// read body till 0
		endChunk := strings.Index(string(bodyB), "\n0\n")
		if endChunk == -1 {
			log.Fatal("Failed to retrieve end of chunks in request('\\n0\\n')")
		}
		bodyTE := string(bodyB[:endChunk+3]) //+3: take into account \n0\n as EndChunk return the index of the substring beginning
		fmt.Printf(bodyTE)

		if len(bodyB) >= endChunk+3 {
			bodyEnding := string(bodyB[endChunk+3:])
			fmt.Fprintf(os.Stderr, utils.Purple(bodyEnding))
		}
	} else {
		fmt.Print(string(bodyB))
	}
}
