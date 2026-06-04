package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	input := "$5\r\nveeya\r\n"                          // just an input string
	reader := bufio.NewReader(strings.NewReader(input)) // string converted to bufio buffer

	b, _ := reader.ReadByte() // data stype that is taken as input ($ is bulk string)
	if b != '$' {
		fmt.Println("invalid type, expexting bulk strings only") //basically smth like $\r\n....\r\n
		os.Exit(1)
	}
	size, _ := reader.ReadByte()                         // size of string
	strsize, _ := strconv.ParseInt(string(size), 10, 64) // converting to int
	reader.ReadByte()                                    // \r
	reader.ReadByte()                                    // \n

	name := make([]byte, strsize) // array of all input string bytes
	reader.Read(name)             // reading it
	fmt.Println(string(name))
	// // a tcp listener
	// l, err := net.Listen("tcp", ":6379")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Listening on :6379")
	// // reciveing requests
	// conn, err := l.Accept()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer conn.Close() // closes connection after finishing (not really required here but is a good practice and will be of good use later)

	// // an infite loop of listening to the requests but
	// for {
	// 	buf := make([]byte, 1024)
	// 	_, err := conn.Read(buf)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		fmt.Println("Error reading from the client: ", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	conn.Write([]byte("+OK\r\n")) // but here just giving back a pong confirmation (demo setup)
	// }
}
