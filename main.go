package main

import (
	"fmt"
	"net"
	"strings"
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

	fmt.Println("Listening on :6379")

	// a tcp listener
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	//aof initialised
	AOF, err := Newaof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer AOF.close() // close when server is shut off

	if err := AOF.read(func(val value) {
		prevcommands := strings.ToUpper(val.array[0].bulk)
		args := val.array[1:]
		handler, ok := handlers[prevcommands]
		if !ok {
			fmt.Println("Invalid command while replaying aof: ", prevcommands)
			return
		}
		handler(args)
	}); err != nil {
		fmt.Println("AOF replay error:", err)
		return
	}
	for {
		// reciveing requests
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleconn(conn, AOF)

	}
}
func handleconn(conn net.Conn, AOF *Aof) {
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
		switch command {
		case "SUBSCRIBE":
			result := subscribe(args, conn)
			writer.write(result)
			continue
		case "UNSUBSCRIBE":
			result := unsubscribe(args, conn)
			writer.write(result)
			continue
		}
		handler, ok := handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.write(value{typ: "string", str: ""})
			continue
		}
		//writing in aof
		if command == "SET" || command == "HSET" {
			AOF.write(val)
		}
		result := handler(args)
		writer.write(result)
		// writer.write(value{typ:"str",str:"ok"})
		// fmt.Println(value)
		// conn.Write([]byte("+OK\r\n")) // but here just giving back a pong confirmation (demo setup)
	}
}
