package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// a tcp listener
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on :6379")
	// reciveing requests
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close() // closes connection after finishing (not really required here but is a good practice and will be of good use later)

	// an infite loop of listening to the requests but
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading from the client: ", err.Error())
			os.Exit(1)
		}

		conn.Write([]byte("+OK\r\n")) // but here just giving back a pong confirmation (demo setup)
	}
}
