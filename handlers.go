package main

import "sync"

var handlers = map[string]func([]value) value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
}

func ping(args []value) value {
	if len(args) == 0 {
		return value{typ: "string", str: "PONG"}
	}
	return value{typ: "string", str: args[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(args []value) value {
	if len(args) != 2 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for a SET command",
		}
	}
	key := args[0].bulk
	val := args[1].bulk
	SETsMu.Lock()
	SETs[key] = val
	SETsMu.Unlock()
	return value{
		typ: "string",
		str: "ok",
	}
}
func get(args []value) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR wrong number of args for a GET command",
		}
	}
	key := args[0].bulk
	SETsMu.RLock()
	val, ok := SETs[key]
	SETsMu.RUnlock()
	if !ok {
		return value{typ: "null"}
	}
	return value{
		typ:  "bulk",
		bulk: val,
	}
}
