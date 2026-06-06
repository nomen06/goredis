package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	// input := "$5\r\nveeya\r\n"                          // just an input string
	// reader := bufio.NewReader(strings.NewReader(input)) // string converted to bufio buffer

	// b, _ := reader.ReadByte() // data stype that is taken as input ($ is bulk string)
	// if b != '$' {
	// 	fmt.Println("invalid type, expexting bulk strings only") //basically smth like $\r\n....\r\n
	// 	os.Exit(1)
	// }
	// size, _ := reader.ReadByte()                         // size of string
	// strsize, _ := strconv.ParseInt(string(size), 10, 64) // converting to int
	// reader.ReadByte()                                    // \r
	// reader.ReadByte()                                    // \n

	// name := make([]byte, strsize) // array of all input string bytes
	// reader.Read(name)             // reading it
	// fmt.Println(string(name))

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
		resp := Newresp(conn)
		val, err := resp.read()
		if err != nil {
			fmt.Println(err)
			return
		}
		if val.typ != "array" {
			fmt.Println("Invalid request,expected array")
			continue
		}
		if len(val.array) == 0 {
			fmt.Println("Invalid request,expected array of len>0")
			continue
		}
		command := strings.ToUpper(val.array[0].bulk)
		args := val.array[1:]

		writer := Newwriter(conn)
		handler, ok := handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.write(value{typ: "string", str: ""})
			continue
		}
		result := handler(args)
		writer.write(result)
		// writer.write(value{typ:"str",str:"ok"})
		// fmt.Println(value)
		// conn.Write([]byte("+OK\r\n")) // but here just giving back a pong confirmation (demo setup)
	}
}
func Newaof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666) //0666->wr-wr-wr , 0644->wr,r,r ,
	if err != nil {
		return nil, err
	}
	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}
	go func() { // a go routine to sync aof to the disc every 1 second
		for {
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()
	return aof, nil
}
