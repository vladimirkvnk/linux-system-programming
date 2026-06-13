package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

const (
	defaultPort = 8080
	maxEvents   = 64
	bufSize     = 4096
)

type Server struct {
	listenFD int
	epollFD int

	events []unix.EpollEvent
	buf    []byte
}

func main() {
	port, err := portFromEnv()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	server, err := NewServer(port)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	defer server.Close()

	log.Printf("Echo server listening on port %d\n", port)

	if err := server.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func NewServer(port int) (*Server, error) {
	listenFD, err := createListenSocket(port)
	if err != nil {
		return nil, err
	}

	epollFD, err := unix.EpollCreate1(0)
	if err != nil {
		closeFD("listen fd", listenFD)
		return nil, fmt.Errorf("epoll_create1: %w", err)
	}

	server := &Server{
		listenFD: listenFD,
		epollFD: epollFD,
		events:  make([]unix.EpollEvent, maxEvents),
		buf:     make([]byte, bufSize),
	}

	if err := server.addReadFD(listenFD); err != nil {
		server.Close()
		return nil, fmt.Errorf("add listen fd to epoll: %w", err)
	}

	return server, nil
}

func (s *Server) Run() error {
	for {
		n, err := unix.EpollWait(s.epollFD, s.events, -1)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}

			return fmt.Errorf("epoll_wait: %w", err)
		}

		s.handleEvents(n)
	}
}

func (s *Server) handleEvents(n int) {
	for i := range n {
		s.handleEvent(s.events[i])
	}
}

func (s *Server) handleEvent(event unix.EpollEvent) {
	fd := int(event.Fd)

	if fd == s.listenFD {
		s.acceptReadyClients()
		return
	}

	if event.Events&(unix.EPOLLERR|unix.EPOLLHUP) != 0 {
		log.Printf("client fd=%d got error/hup event", fd)
		s.closeClient(fd)
		return
	}

	if event.Events&unix.EPOLLIN != 0 {
		s.handleClientReadable(fd)
	}
}

func (s *Server) acceptReadyClients() {
	for {
		clientFD, clientAddr, err := unix.Accept(s.listenFD)
		if err != nil {
			if isWouldBlock(err) {
				return
			}

			log.Printf("accept: %v", err)
			return
		}

		if err := s.addClient(clientFD); err != nil {
			log.Printf("failed to add client fd=%d: %v", clientFD, err)
			closeFD("client fd", clientFD)
			continue
		}

		log.Printf("Client connected: fd=%d, addr=%v\n", clientFD, clientAddr)
	}
}

func (s *Server) addClient(clientFD int) error {
	if err := unix.SetNonblock(clientFD, true); err != nil {
		return fmt.Errorf("set nonblocking: %w", err)
	}

	if err := s.addReadFD(clientFD); err != nil {
		return fmt.Errorf("add to epoll: %w", err)
	}

	return nil
}

func (s *Server) handleClientReadable(fd int) {
	for {
		bytesRead, err := unix.Read(fd, s.buf)
		if err != nil {
			if isWouldBlock(err) {
				return
			}

			log.Printf("read fd=%d: %v", fd, err)
			s.closeClient(fd)
			return
		}

		if bytesRead == 0 {
			log.Printf("Client disconnected: fd=%d\n", fd)
			s.closeClient(fd)
			return
		}

		if err := writeAll(fd, s.buf[:bytesRead]); err != nil {
			log.Printf("write fd=%d: %v", fd, err)
			s.closeClient(fd)
			return
		}
	}
}

func (s *Server) addReadFD(fd int) error {
	event := unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(fd),
	}

	return unix.EpollCtl(s.epollFD, unix.EPOLL_CTL_ADD, fd, &event)
}

func (s *Server) closeClient(fd int) {
	_ = unix.EpollCtl(s.epollFD, unix.EPOLL_CTL_DEL, fd, nil)
	closeFD("client fd", fd)
}

func (s *Server) Close() {
	closeFD("listen fd", s.listenFD)
	closeFD("epoll fd", s.epollFD)
}

func createListenSocket(port int) (int, error) {
	listenFD, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		return -1, fmt.Errorf("socket: %w", err)
	}

	success := false
	defer func() {
		if !success {
			closeFD("listen fd", listenFD)
		}
	}()

	if err := unix.SetsockoptInt(listenFD, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return -1, fmt.Errorf("setsockopt SO_REUSEADDR: %w", err)
	}

	if err := unix.SetNonblock(listenFD, true); err != nil {
		return -1, fmt.Errorf("set nonblocking: %w", err)
	}

	addr := unix.SockaddrInet4{
		Port: port,
		Addr: [4]byte{0, 0, 0, 0},
	}

	if err := unix.Bind(listenFD, &addr); err != nil {
		return -1, fmt.Errorf("bind: %w", err)
	}

	if err := unix.Listen(listenFD, unix.SOMAXCONN); err != nil {
		return -1, fmt.Errorf("listen: %w", err)
	}

	success = true
	return listenFD, nil
}

func writeAll(fd int, data []byte) error {
	for len(data) > 0 {
		n, err := unix.Write(fd, data)
		if err != nil {
			return err
		}

		if n == 0 {
			return errors.New("write returned 0 bytes")
		}

		data = data[n:]
	}

	return nil
}

func isWouldBlock(err error) bool {
	return errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK)
}

func closeFD(name string, fd int) {
	if fd < 0 {
		return
	}

	if err := unix.Close(fd); err != nil {
		log.Printf("failed to close %s: %v", name, err)
	}
}

func portFromEnv() (int, error) {
	value := os.Getenv("PORT")
	if value == "" {
		return defaultPort, nil
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("PORT must be a number: %w", err)
	}

	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("PORT must be in range 1..65535, got %d", port)
	}

	return port, nil
}
