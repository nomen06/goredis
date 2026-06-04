package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []value
}
type Resp struct {
	reader *bufio.Reader
}

func Newresp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readline() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) > 1 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *Resp) readint() (x int, n int, err error) {
	line, n, err := r.readline()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) read() (value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return value{}, err
	}
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type :%v", string(_type))
		return value{}, nil
	}
}
func (r *Resp) readArray() (value, error) {
	v := value{}
	v.typ = "array"

	length, _, err := r.readint()
	if err != nil {
		return v, err
	}
	v.array = make([]value, length)
	for i := 0; i < length; i++ {
		val, err := r.read()
		if err != nil {
			return v, err
		}
		v.array[i] = val
	}
	return v, nil
}
func (r *Resp) readBulk() (value, error) {
	v := value{}
	v.typ = "bulk"
	len, _, err := r.readint()
	if err != nil {
		return v, err
	}
	bulk := make([]byte, len)
	r.reader.Read(bulk)
	v.bulk = string(bulk)
	r.readline()
	return v, nil
}

// write RESP

type Writer struct {
	writer io.Writer
}

func Newwriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) write(v value) error {
	var bytes = v.marshal()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (v value) marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalarray()
	case "bulk":
		return v.marshalbulk()
	case "string":
		return v.marshalstring()
	case "null":
		return v.marshalnull()
	case "error":
		return v.marshalerror()
	default:
		return []byte{}
	}
}
func (v value) marshalstring() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}
func (v value) marshalbulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}
func (v value) marshalarray() []byte {
	length := len(v.array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, '\r', '\n')
	for i := 0; i < length; i++ {
		bytes = append(bytes, v.array[i].marshal()...)
	}
	return bytes
}
func (v value) marshalerror() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}
func (v value) marshalnull() []byte {
	return []byte("$-1\r\n")
}
