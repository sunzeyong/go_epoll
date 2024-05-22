package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Println("server is starting...")

	netPoll := NewNetPoll("10.12.100.212", 8639)

	if err := netPoll.CreateEpoll(); err != nil {
		log.Fatalf("fail to create epoll, err:%s", err)
	}

	if err := netPoll.InitListenFd(); err != nil {
		log.Fatalf("fail to listen, err:%s", err)
	}

	go func() {
		if err := netPoll.EventHandler(); err != nil {
			defer netPoll.Close()
			log.Fatalf("fail to exec event handler, err:%s", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	netPoll.Close()
	log.Println("exit successfully")
}
