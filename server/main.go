package main

import "log"

func main() {
	log.Println("server is starting...")

	netPoll := NewNetPoll()

	err := netPoll.CreateEpoll()
	if err != nil {
		panic(err)
	}

	err = netPoll.CreateListenFd("0.0.0.0", 8088)
	if err != nil {
		panic(err)
	}

	go func() {
		err := netPoll.EventHandler()
		netPoll.Close()
		panic(err)
	}()

	err = netPoll.Accept()
	netPoll.Close()
	panic(err)
}
