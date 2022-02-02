package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"HTTPCustomHouse/pkg/parser"
	"HTTPCustomHouse/pkg/utils"
)

const usage = `Usage of customOfficer-CL:
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
	bodyCL := string(bodyB[:contentLength-1]) // -1 due to the \n beginning the body form (see above)
	fmt.Printf(bodyCL)

	// Print request residue
	if residue && len(bodyB) >= contentLength {
		bodyResidue := string(bodyB[contentLength-1:])
		fmt.Fprintf(os.Stderr, utils.Purple(bodyResidue))
	}

}
