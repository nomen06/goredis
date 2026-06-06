package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
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

// to ensure aof is closed propwrly after server shuts down
func (aof *Aof) close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	return aof.file.Close()
}

//write method for commands in aof

func (aof *Aof) write(val value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	_, err := aof.file.Write(val.marshal())
	if err != nil {
		return err
	}
	return nil
}

// reading aof file (assuming not corrupted yet JUST AN EXAMPLE)
// read method
func (aof *Aof) read(callback func(val value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	resp := Newresp(aof.file)
	for {
		val, err := resp.read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		callback(val)
	}
	return nil
}
