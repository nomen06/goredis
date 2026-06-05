package main

var handlers = map[string]func([]value) value{
	"PING": ping,
}

func ping(args []value) value {
	if len(args) == 0 {
		return value{typ: "string", str: "PONG"}
	}
	return value{typ: "string", str: args[0].bulk}
}
