package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
)

// To test your server implementation, you might find it helpful to implement a
// simple 'client runner' program. The program could be very simple, as long as
// it is able to connect with and send messages to your server and is able to
// read and print out the server's echoed response to standard output. Whether or
// not you add any code to this file will not affect your grade.
func main() {
	fmt.Println("Not implemented.")
	conn, err := net.Dial("tcp", defaultHost+":"+strconv.Itoa(defaultPort))
	if err != nil {
		fmt.Println("error dialing...", err)
		return
	}
	inputReader := bufio.NewReader(os.Stdin)
	go read(conn)
	for {
		fmt.Println("What to send to the server? Type Q to quit.")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\n")
		if trimmedInput == "Q" {
			return
		}
		_, err = conn.Write([]byte(input))
	}

}
func read(conn net.Conn) {
	for {
		buf := make([]byte, 512)
		len, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading...", err.Error())
			//delete

			return
		}
		fmt.Printf("Received data: %v", string(buf[:len]))
	}
}
