package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/nbright/nomadcoin/explorer"
	"github.com/nbright/nomadcoin/rest"
)

func usage() {
	fmt.Printf("Welcome to 노마드코인 \n\n")
	fmt.Printf("Please use following commands: \n\n")
	fmt.Printf("-port:     	Start the REST API (recommanded) \n")
	fmt.Printf("-mode: 		Choose between 'html' and 'rest' \n")
	runtime.Goexit()
}
func Start() {
	if len(os.Args) == 1 {
		usage()
	}

	port := flag.Int("port", 4000, "Set port of the server")
	mode := flag.String("mode", "rest", "Choose between 'html' and 'rest'")

	flag.Parse()

	switch *mode {
	case "rest":
		rest.Start(*port)
	case "html":
		explorer.Start(*port)
	default:
		usage()
	}

}

func flagSetMain() {
	rest := flag.NewFlagSet("rest", flag.ExitOnError)
	portFlag := rest.Int("port", 4000, "Sets the port of server (default 4000)")

	switch os.Args[1] {
	case "rest":
		rest.Parse(os.Args[2:])
	case "explorer":
		fmt.Println("rest")
	default:
		usage()

	}
	if rest.Parsed() {
		fmt.Println(*portFlag)
		fmt.Println("Start Server")
	}
}
