package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func formatresp(args []string) string {
	n := len(args)
	input := "*"
	input += strconv.Itoa(n)
	input += "\r"
	input += "\n"
	for comp := range args {
		len2 := len(args[comp])
		input += "$"
		input += strconv.Itoa(len2)
		input += "\r"
		input += "\n"
		input += args[comp]
		input += "\r"
		input += "\n"
	}
	return input
}
func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	Reader := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)
	for {
		input, err := Reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		args := strings.Fields(input)
		value := formatresp(args)
		fmt.Fprint(conn, value)

		response, err := connReader.ReadString('\n')
		fmt.Fprint(os.Stdout, response)
	}

}
