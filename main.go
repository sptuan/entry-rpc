package main

import (
	"entry-rpc/internal/client"
	"entry-rpc/internal/server"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	server.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	clt, _ := client.Dial("tcp", <-addr)
	defer func() { _ = clt.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("geerpc req %d", i)
			var reply string
			for j := 0; j < 1000000; j++ {
				if err := clt.Call("Foo.Sum", args, &reply); err != nil {
					log.Fatal("call Foo.Sum error:", err)
				}
				log.Println("reply:", reply)
			}
		}(i)
	}
	wg.Wait()
}
