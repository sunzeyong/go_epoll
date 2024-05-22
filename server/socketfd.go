package main

import (
	"fmt"
	"log"
	"syscall"
)

type SocketFd struct {
	fd         int
	isListenFd bool
}

func (s *SocketFd) Handler(netpoll *NetPoll) error {
	if s.isListenFd {
		nfd, _, err := syscall.Accept(s.fd)
		if err != nil {
			return err
		}
		err = netpoll.AddEpollItem(nfd)
		if err != nil {
			return nil
		}
		netpoll.AddConn(&SocketFd{
			fd: nfd,
		})
	} else {
		s.Read()
	}

	return nil
}

func (s *SocketFd) Read() {
	// 最大读取1KB数据
	data := make([]byte, 1024)

	n, err := syscall.Read(s.fd, data)
	if n == 0 {
		log.Printf("socket fd:%v, read zero byte\n", s.fd)
		return
	}
	if err != nil {
		log.Printf("fd %d read error:%s\n", s.fd, err.Error())
	}

	readInfo := data[:n]
	log.Printf("socket fd %d says: %s \n", s.fd, readInfo)
	s.Write([]byte(s.GetResp(string(readInfo))))
}

func (s *SocketFd) Write(data []byte) {
	_, err := syscall.Write(s.fd, data)
	if err != nil {
		log.Printf("fd %d write error:%s\n", s.fd, err.Error())
	}
}

func (s *SocketFd) Close() {
	err := syscall.Close(s.fd)
	if err != nil {
		log.Printf("fd %d close error:%s\n", s.fd, err.Error())
	}
}

func (s *SocketFd) GetResp(input string) string {
	switch input {
	case "hello":
		return fmt.Sprintf("hello %d", s.fd)
	case "what's weather today?":
		return "it's sunny"
	}
	return "can't find resp"
}
