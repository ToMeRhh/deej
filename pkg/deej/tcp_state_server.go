package deej

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"go.uber.org/zap"
)

type TCPStateServer struct {
	deej   *Deej
	logger *zap.SugaredLogger

	listener net.Listener
}

func NewTCPStateServer(deej *Deej, logger *zap.SugaredLogger) (*TCPStateServer, error) {
	logger = logger.Named("tcp_state_server")

	s, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", deej.config.TcpStateServerConnectionInfo.TcpPort))
	if err != nil {
		logger.Warnw("Failed to resolve tcp address", "error", err)
		return nil, fmt.Errorf("resolve tcp address: %w", err)
	}

	ln, err := net.Listen("tcp4", s.String()) // Listen on all interfaces on the given port
	if err != nil {
		logger.Warnw("Failed to resolve tcp address", "error", err)
		return nil, fmt.Errorf("error creating TCP server: %w", err)
	}

	tcp_state_server := &TCPStateServer{
		deej:     deej,
		logger:   logger,
		listener: ln,
	}

	logger.Debugw("Created TCP State server:", "address", ln.Addr())

	return tcp_state_server, nil
}

func (s *TCPStateServer) Start() error {
	s.logger.Info("Server started. Listening on", s.listener.Addr())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue // Handle the error, but keep the server running
		}
		go s.handleConnection(conn) // Handle connections concurrently
	}
}

func (s *TCPStateServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Info("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n') // Read the request (up to newline)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	request = strings.TrimSpace(request) // Remove leading/trailing whitespace

	if strings.Contains(request, "get_state") {
		_, err := conn.Write([]byte("Hello World\n"))
		if err != nil {
			log.Println("Error writing to connection:", err)
		} else {
			s.logger.Info("Message sent to client:", conn.RemoteAddr())
		}
	} else {
		s.logger.Info("Request does not contain 'get_state':", request)
		_, err := conn.Write([]byte("Invalid request\n"))
		if err != nil {
			log.Println("Error writing error message:", err)
		}
	}
}

func (s *TCPStateServer) Stop() error {
	s.logger.Info("Server shutting down...")
	return s.listener.Close()
}
