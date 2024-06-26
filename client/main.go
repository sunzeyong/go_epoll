package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	ConcurrentClients()

}

func PingOnce() {
	conn, err := net.Dial("tcp", "10.12.100.212:8639")
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("what's weather today?"))
	data := make([]byte, 1000)
	n, err := conn.Read(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data[:n]))

	time.Sleep(100 * time.Millisecond)
	conn.Close()
}

func ConcurrentClients() {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.Dial("tcp", "10.12.100.212:8639")
			if err != nil {
				panic(err)
			}
			conn.Write([]byte("hello"))
			data := make([]byte, 100)
			n, err := conn.Read(data)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data[:n]))

			time.Sleep(100 * time.Millisecond)
			conn.Close()
		}()
	}
	wg.Wait()
}
