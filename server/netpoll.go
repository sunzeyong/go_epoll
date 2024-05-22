package main

import (
	"log"
	"net"
	"sync"
	"syscall"
)

type NetPoll struct {
	locker    sync.RWMutex
	socketMap map[int]*SocketFd

	backlog int
	ip      string
	port    int

	epollFd int //epoll的fd
}

func NewNetPoll(ip string, port int) *NetPoll {
	return &NetPoll{
		backlog: 200,
		ip:      ip,
		port:    port,

		socketMap: make(map[int]*SocketFd),
	}
}

func (e *NetPoll) CreateEpoll() error {
	// epollCreate的入参表示初始化容量大小，在Linux环境下该值无效
	fd, err := syscall.EpollCreate(1)
	if err != nil {
		return err
	}
	e.epollFd = fd
	log.Printf("create epoll fd successfully, %v\n", fd)

	return nil
}

// 建立listen socket
// 将listen socket添加到epoll中
func (e *NetPoll) InitListenFd() error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}

	if err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Addr: [4]byte(net.ParseIP(e.ip).To4()),
		Port: e.port,
	}); err != nil {
		return err
	}

	if err := syscall.Listen(fd, e.backlog); err != nil {
		return err
	}

	e.AddListenEpollItem(fd)
	e.AddConn(&SocketFd{fd, true})
	log.Printf("create listen fd successfully, %v\n", fd)

	return nil
}

func (e *NetPoll) EventHandler() error {
	events := make([]syscall.EpollEvent, 1024)
	for {
		log.Printf("socketmap: %v\n", e.socketMap)

		n, err := syscall.EpollWait(e.epollFd, events, -1)
		log.Printf("event:%v\n", events[:n])

		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			conn := e.GetConn(int(events[i].Fd))
			if conn == nil {
				continue
			}
			if events[i].Events&syscall.EPOLLRDHUP == syscall.EPOLLRDHUP || events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR {
				if err := e.CloseConn(int(events[i].Fd)); err != nil {
					return err
				}
				log.Printf("close conn successfully, fd: %v", conn.fd)
			}
			if events[i].Events == syscall.EPOLLIN {
				if err := conn.Handler(e); err != nil {
					return err
				}
			}
		}
	}
}

func (e *NetPoll) Close() {
	for _, con := range e.socketMap {
		con.Close()
	}

	syscall.Close(e.epollFd)
}

func (e *NetPoll) GetConn(fd int) *SocketFd {
	e.locker.RLock()
	defer e.locker.RUnlock()

	return e.socketMap[fd]
}

func (e *NetPoll) AddConn(conn *SocketFd) {
	e.locker.Lock()
	defer e.locker.Unlock()

	e.socketMap[conn.fd] = conn
}

func (e *NetPoll) DelConn(fd int) {
	e.locker.Lock()
	defer e.locker.Unlock()

	delete(e.socketMap, fd)
}

func (e *NetPoll) CloseConn(fd int) error {
	conn := e.GetConn(fd)
	if conn == nil {
		return nil
	}
	if err := e.RemoveEpollItem(fd); err != nil {
		return err
	}
	conn.Close()
	e.DelConn(fd)
	return nil
}

// listen使用水平触发，只要有数据就继续处理
func (e *NetPoll) AddListenEpollItem(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
		Pad:    0,
	})
}

// connect socket使用
func (e *NetPoll) AddEpollItem(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLLRDHUP,
		Fd:     int32(fd),
		Pad:    0,
	})
}

func (e *NetPoll) RemoveEpollItem(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}
