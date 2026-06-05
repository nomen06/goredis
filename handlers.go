package main

import "sync"

var handlers = map[string]func([]value) value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
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

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func hset(args []value) value {
	if len(args) != 3 {
		return value{
			typ: "error",
			str: "ERR hset command requires 3 arguments",
		}
	}
	hash := args[0].bulk
	key := args[1].bulk
	val := args[2].bulk
	HSETsMu.Lock()
	_, ok := HSETs[hash]
	if !ok {
		HSETs[hash] = map[string]string{}
	}
	HSETs[hash][key] = val
	HSETsMu.Unlock()
	return value{typ: "string", str: "ok"}
}

func hget(args []value) value {
	if len(args) != 2 {
		return value{typ: "error", str: "ERR hget require 2 number of args"}
	}
	hash := args[0].bulk
	key := args[1].bulk
	HSETsMu.RLock()
	val, ok := HSETs[hash][key]
	HSETsMu.RUnlock()
	if !ok {
		return value{typ: "null"}
	}
	return value{
		typ:  "bulk",
		bulk: val,
	}
}
func hgetall(args []value) value {
	if len(args) != 1 {
		return value{typ: "error", str: "ERR hgetall requires only 1 argument"}
	}
	hash := args[0].bulk
	HSETsMu.RLock()
	inner, ok := HSETs[hash]
	HSETsMu.RUnlock()
	if !ok {
		return value{typ: "null"}
	}
	var result []value
	for k, v := range inner {
		result = append(result, value{typ: "bulk", bulk: k})
		result = append(result, value{typ: "bulk", bulk: v})
	}
	return value{typ: "array", array: result}
}
