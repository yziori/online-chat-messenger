package handlers

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/yziori/online-chat-messenger/internal/models"
)

type Server struct {
	Addr    string
	Clients map[string]*models.Client
	Mutex   sync.Mutex
}

func NewServer(addr string) *Server {
	return &Server{
		Addr:    addr,
		Clients: make(map[string]*models.Client),
	}
}

func (s *Server) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buffer := make([]byte, 4096)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Read error from UDP: ", err)
			continue
		}

		go s.handleMessage(conn, addr, buffer[:n])
	}
}

func (s *Server) handleMessage(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	usernameSize := int(msg[0])
	username := string(msg[1 : 1+usernameSize])
	message := string(msg[1+usernameSize:])

	client, exists := s.Clients[addr.String()]
	if !exists {
		client = &models.Client{
			Addr:       addr,
			Username:   username,
			LastActive: time.Now(),
		}
		s.Clients[addr.String()] = client
	} else {
		client.LastActive = time.Now()
	}

	fmt.Printf("[%s] %s: %s\n", addr.String(), username, message)

	s.relayMessage(conn, addr, msg)
}

func (s *Server) relayMessage(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) {
	for _, client := range s.Clients {
		if client.Addr.String() != addr.String() {
			conn.WriteToUDP(msg, client.Addr)
		}
	}
}
