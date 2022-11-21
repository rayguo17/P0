package main

import (
	"bufio"
	"fmt"
	"github.com/cmu440/p0partA"
	"github.com/cmu440/p0partA/kvstore"
	"github.com/k0kubun/pp/v3"
	"os"
)

const defaultPort = 9999

func main() {
	// Initialize the key-value store.
	store, _ := kvstore.CreateWithBackdoor()
	// Initialize the server.
	server := p0partA.New(store)
	if server == nil {
		fmt.Println("New() returned a nil server. Exiting...")
		return
	}

	// Start the server and continue listening for client connections in the background.
	if err := server.Start(defaultPort); err != nil {
		fmt.Printf("KeyValueServer could not be started: %s\n", err)
		return
	}

	fmt.Printf("Started KeyValueServer on port %d...\n", defaultPort)
	inputReader := bufio.NewReader(os.Stdin)
	for {
		str, _ := inputReader.ReadString('\n')
		switch str {
		case "1\n":
			pp.Println(store)
		case "2\n":
			pp.Println(server)
		case "3\n":
			server.Close()
			return
		}

	}
	// Block forever.
}
