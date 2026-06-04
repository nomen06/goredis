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
