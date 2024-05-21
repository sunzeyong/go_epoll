package main

import (
	"log"
	"net"
	"sync"
	"syscall"
)

type NetPoll struct {
	locker sync.RWMutex
	conns  map[int]*SocketFd

	listenFd int //监听socket的fd
	epollFd  int //epoll的fd
}

func NewNetPoll() *NetPoll {
	return &NetPoll{
		conns: make(map[int]*SocketFd),
	}
}

func (e *NetPoll) Close() {
	syscall.Close(e.listenFd)
	syscall.Close(e.epollFd)

	for _, con := range e.conns {
		con.Close()
	}
}

func (e *NetPoll) GetConn(fd int) *SocketFd {
	e.locker.RLock()
	defer e.locker.RUnlock()

	return e.conns[fd]
}

func (e *NetPoll) AddConn(conn *SocketFd) {
	e.locker.Lock()
	defer e.locker.Unlock()

	e.conns[conn.fd] = conn
}

func (e *NetPoll) DelConn(fd int) {
	e.locker.Lock()
	defer e.locker.Unlock()

	delete(e.conns, fd)
}

func (e *NetPoll) CreateListenFd(ipAddr string, port int) error {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}

	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Addr: [4]byte(net.ParseIP(ipAddr).To4()),
		Port: port,
	})
	if err != nil {
		return err
	}

	err = syscall.Listen(fd, 10)
	if err != nil {
		return err
	}
	e.listenFd = fd
	log.Printf("create listen fd successfully, %v\n", fd)
	return nil
}

func (e *NetPoll) Accept() error {
	for {
		nfd, _, err := syscall.Accept(e.listenFd)
		if err != nil {
			return err
		}
		err = e.EpollAddEvent(nfd)
		if err != nil {
			return nil
		}
		e.AddConn(&SocketFd{
			fd: nfd,
		})
	}
}

func (e *NetPoll) CloseConn(fd int) error {
	conn := e.GetConn(fd)
	if conn == nil {
		return nil
	}
	if err := e.EpollRemoveEvent(fd); err != nil {
		return err
	}
	conn.Close()
	e.DelConn(fd)
	return nil
}

func (e *NetPoll) CreateEpoll() error {
	fd, err := syscall.EpollCreate(1)
	if err != nil {
		return err
	}
	e.epollFd = fd
	log.Printf("create epoll fd successfully, %v\n", fd)

	return nil
}

func (e *NetPoll) EventHandler() error {
	events := make([]syscall.EpollEvent, 100)
	for {
		n, err := syscall.EpollWait(e.epollFd, events, -1)

		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			conn := e.GetConn(int(events[i].Fd))
			if conn == nil {
				continue
			}
			if events[i].Events&syscall.EPOLLHUP == syscall.EPOLLHUP || events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR {
				if err := e.CloseConn(int(events[i].Fd)); err != nil {
					return err
				}
			} else if events[i].Events == syscall.EPOLLIN {
				conn.Read()
			}
		}
	}
}

func (e *NetPoll) EpollAddEvent(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
		Pad:    0,
	})
}

func (e *NetPoll) EpollRemoveEvent(fd int) error {
	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}
