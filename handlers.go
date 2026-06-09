package main

import (
	"strconv"
	"sync"
	"time"
)

var handlers = map[string]func([]value) value{
	"PING":      ping,
	"SET":       set,
	"GET":       get,
	"HSET":      hset,
	"HGET":      hget,
	"HGETALL":   hgetall,
	"EXPIRE":    expire,
	"TTL":       ttl,
	"SUBSCRIBE": subscribe,
	"PUBLISH":   publish,
}

func publish(args []value) value {
	if len(args) != 2 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for a PUBLISH command",
		}
	}
	channel := args[0].bulk
	message := args[1].bulk
	SUBSCRIBEsMu.RLock()
	_, check := SUBSCRIBEs[channel]
	if !check {
		return value{
			typ: "integer",
			num: 0,
		}
	}
	SUBSCRIBEsMu.RUnlock()
	for _, ch := range SUBSCRIBEs[channel] {
		go func(ch chan string) { ch <- message }(ch)
	}
	return value{
		typ:  "bulk",
		bulk: message,
	}
}

var SUBSCRIBEs = map[string][]chan string{} // [] because multiple users can listen to a single channel
var SUBSCRIBEsMu = sync.RWMutex{}

func subscribe(args []value) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for a SUBSCRIBE command",
		}
	}
	channel := args[0].bulk
	ch := make(chan string)
	SUBSCRIBEsMu.Lock()
	SUBSCRIBEs[channel] = append(SUBSCRIBEs[channel], ch) // this makes that multiple user thing make happen
	SUBSCRIBEsMu.Unlock()
	msg := <-ch
	return value{
		typ: "string",
		str: msg,
	}
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
	EXPIREsMu.Lock()
	expiry, expcheck := EXPIREs[key]
	if expcheck {
		if time.Now().After(expiry) {
			SETsMu.Lock()
			delete(SETs, key)
			SETsMu.Unlock()
			delete(EXPIREs, key)
			EXPIREsMu.Unlock()
			return value{
				typ: "null",
			}
		}
	}
	EXPIREsMu.Unlock()
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

var EXPIREs = map[string]time.Time{}
var EXPIREsMu = sync.RWMutex{}

func expire(args []value) value {
	if len(args) != 2 {
		return value{typ: "error", str: "ERR,invalid number of arguments for EXPIRE command"}
	}
	key := args[0].bulk
	val := args[1].bulk
	SETsMu.RLock()
	_, ok := SETs[key]
	SETsMu.RUnlock()
	if !ok {
		return value{
			typ: "integer",
			num: 0,
		}
	}
	secs, err := strconv.Atoi(val)
	if err != nil {
		return value{
			typ: "error",
			str: "ERR value is not an integer",
		}
	}
	deadline := time.Now().Add(time.Duration(secs) * time.Second)
	EXPIREsMu.Lock()
	EXPIREs[key] = deadline
	EXPIREsMu.Unlock()
	return value{
		typ: "integer",
		num: 1,
	}
}

func ttl(args []value) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR invalid number of args for ttl",
		}
	}
	key := args[0].bulk
	SETsMu.RLock()
	_, ok := SETs[key]
	SETsMu.RUnlock()
	if !ok {
		return value{
			typ: "integer",
			num: -2,
		}
	}
	EXPIREsMu.Lock()
	expiry, check := EXPIREs[key]
	if !check {
		EXPIREsMu.Unlock()
		return value{
			typ: "integer",
			num: -1,
		}
	}
	if time.Now().After(expiry) {
		SETsMu.Lock()
		delete(SETs, key)
		SETsMu.Unlock()
		delete(EXPIREs, key)
		EXPIREsMu.Unlock()
		return value{
			typ: "null",
		}
	}
	ttl := time.Until(expiry)
	EXPIREsMu.Unlock()
	return value{
		typ: "integer",
		num: int(ttl.Seconds()),
	}
}
