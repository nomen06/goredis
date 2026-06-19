package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func formatresp(args []string) string {
	n := len(args)
	input := "*"
	input += strconv.Itoa(n)
	input += "\r\n"
	for _, a := range args {
		input += "$"
		input += strconv.Itoa(len(a))
		input += "\r\n"
		input += a
		input += "\r\n"
	}
	return input
}

func main() {
	numSubscribers := 10
	numMessages := 100

	var wg sync.WaitGroup
	var receivedCount int
	var mu sync.Mutex
	ready := make(chan bool, numSubscribers)

	// spawning sample subscribers to check it's speeeeddd
	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", "localhost:6379")
			if err != nil {
				fmt.Println("subscriber connect error:", err)
				return
			}
			defer conn.Close()
			reader := bufio.NewReader(conn)
			fmt.Fprint(conn, formatresp([]string{"SUBSCRIBE", "bench"}))
			ready <- true

			for j := 0; j < numMessages; j++ {
				_, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				mu.Lock()
				receivedCount++
				mu.Unlock()
			}
		}()
	}

	// waiting for all subscribers to confirm they've subscribed
	for i := 0; i < numSubscribers; i++ {
		<-ready
	}
	time.Sleep(200 * time.Millisecond) // a lil bufffer time

	start := time.Now()

	pubConn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("publisher connect error:", err)
		return
	}
	defer pubConn.Close()
	pubReader := bufio.NewReader(pubConn)

	for i := 0; i < numMessages; i++ {
		fmt.Fprint(pubConn, formatresp([]string{"PUBLISH", "bench", "hello"}))
		pubReader.ReadString('\n')
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := numSubscribers * numMessages
	fmt.Printf("Subscribers: %d | Messages published: %d\n", numSubscribers, numMessages)
	fmt.Printf("Total messages delivered: %d in %v\n", receivedCount, elapsed)
	fmt.Printf("Throughput: %.0f messages/sec\n", float64(total)/elapsed.Seconds())
}
