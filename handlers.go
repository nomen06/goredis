package main

import (
	"net"
	"strconv"
	"sync"
	"time"
)

var handlers = map[string]func([]value) value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
	"EXPIRE":  expire,
	"TTL":     ttl,
	"PUBLISH": publish,
	"INCR":    increment,
	"SETEX":   setex,
	"DEL":     del,
}

func del(args []value) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for a DEL command",
		}
	}
	key := args[0].bulk
	SETsMu.Lock()
	defer SETsMu.Unlock()
	EXPIREsMu.Lock()
	defer EXPIREsMu.Unlock()
	_, ok := SETs[key]
	if !ok {
		return value{
			typ: "integer",
			num: 0,
		}
	}
	delete(SETs, key)
	delete(EXPIREs, key)
	return value{
		typ: "integer",
		num: 1,
	}
}
func setex(args []value) value {
	if len(args) != 3 {
		return value{
			typ: "error",
			str: "ERR worng number of arguments for a SETEX command",
		}
	}
	key := args[0].bulk
	seconds, err := strconv.Atoi(args[1].bulk)
	val := args[2].bulk
	if err != nil {
		return value{
			typ: "error",
			str: "ERR in number of seconds",
		}
	}
	SETsMu.Lock()
	EXPIREsMu.Lock()
	defer SETsMu.Unlock()
	defer EXPIREsMu.Unlock()
	SETs[key] = val
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	EXPIREs[key] = deadline
	return value{
		typ: "string",
		str: "ok",
	}
}
func increment(args []value) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for an INCR command",
		}
	}
	key := args[0].bulk
	SETsMu.Lock()
	defer SETsMu.Unlock()
	val, err := strconv.Atoi(SETs[key])
	if err != nil && SETs[key] != "" {
		return value{
			typ: "error",
			str: "The key contains a value of the wrong type",
		}
	}
	// if SETs[key]=="" {
	// 	val=0
	// }
	// don't need this block because if this is true the output is already 0,nil

	SETs[key] = strconv.Itoa(val + 1)
	return value{
		typ: "integer",
		num: val + 1,
	}
}

func unsubscribe(args []value, conn net.Conn) value {
	if len(args) != 1 {
		return value{
			typ: "error",
			str: "ERR wrong number of arguments for an UNSUBSCRIBE command",
		}
	}
	channel := args[0].bulk

	SUBSCRIBEsMu.Lock()
	defer SUBSCRIBEsMu.Unlock()
	subscribers, check := SUBSCRIBEs[channel]
	if !check {
		return value{
			typ: "integer",
			num: 0,
		}
	}

	for i, ch := range subscribers {

		if ch.conn == conn {
			close(ch.ch)
			SUBSCRIBEs[channel] = append(subscribers[:i], subscribers[i+1:]...)
			return value{
				typ: "array",
				array: []value{
					{typ: "bulk", bulk: "unsubscribe"},
					{typ: "bulk", bulk: channel},
					{typ: "integer", num: len(SUBSCRIBEs[channel])},
				},
			}
		}
	}
	// not found case
	return value{
		typ: "array",
		array: []value{
			{typ: "bulk", bulk: "unsubscribe"},
			{typ: "bulk", bulk: channel},
			{typ: "integer", num: 0},
		},
	}
}

type Subscriber struct {
	conn net.Conn
	ch   chan string
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
		SUBSCRIBEsMu.RUnlock()
		return value{
			typ: "integer",
			num: 0,
		}
	}
	for _, ch := range SUBSCRIBEs[channel] {
		go func(ch chan string) { ch <- message }(ch.ch) // ch may change overtime so passed ch as an arg in the func
	}
	SUBSCRIBEsMu.RUnlock()
	count := len(SUBSCRIBEs[channel])
	return value{
		typ: "integer",
		num: count,
	}
}

// var SUBSCRIBEs = map[string][]chan string{} // [] because multiple users can listen to a single channel
// so i cannot simply store it like above because now i need to implement unsubscribe and for that i need somtehing that is unique about all the clients
// since that is the connection between them, hence

var SUBSCRIBEs = map[string][]Subscriber{}
var SUBSCRIBEsMu = sync.RWMutex{}

func subscribe(args []value, conn net.Conn) {
	if len(args) != 1 {
		writer := Newwriter(conn)
		writer.write(value{
			typ: "error",
			str: "ERR wrong number of arguments for a SUBSCRIBE command",
		})
		return
	}
	channel := args[0].bulk
	ch := make(chan string)
	sub := Subscriber{
		conn: conn,
		ch:   ch,
	}
	SUBSCRIBEsMu.Lock()
	SUBSCRIBEs[channel] = append(SUBSCRIBEs[channel], sub) // this makes that multiple user thing make happen
	SUBSCRIBEsMu.Unlock()
	// go func() {
	// 	writer := Newwriter(conn)
	// 	for msg := range ch {
	// 		writer.write(value{
	// 			typ: "string",
	// 			str: msg,
	// 		})
	// 	}
	// }() // this was causing immediate returning

	writer := Newwriter(conn)
	writer.write(value{
		typ: "array",
		array: []value{
			{typ: "bulk", bulk: "subscribe"},
			{typ: "bulk", bulk: channel},
			{typ: "integer", num: 1},
		},
	})
	go func() {
		for msg := range ch {
			writer.write(value{
				typ: "array",
				array: []value{
					{typ: "bulk", bulk: "message"},
					{typ: "bulk", bulk: channel},
					{typ: "bulk", bulk: msg},
				},
			})
		}
		// this part was returning after a single msg
		// msg := <-ch
		// return value{
		// 	typ: "string",
		// 	str: msg,
		// }
	}()
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

func cleanup(conn net.Conn) {
	SUBSCRIBEsMu.Lock()
	defer SUBSCRIBEsMu.Unlock()

	for channel, subscribers := range SUBSCRIBEs {
		var remaining []Subscriber
		for _, sub := range subscribers {
			if sub.conn != conn {
				remaining = append(remaining, sub)
			} else {
				close(sub.ch)
			}
		}
		if len(remaining) == 0 {
			delete(SUBSCRIBEs, channel)
		} else {
			SUBSCRIBEs[channel] = remaining
		}
	}
}
