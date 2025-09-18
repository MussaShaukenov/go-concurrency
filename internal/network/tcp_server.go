package network

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/MussaShaukenov/go-concurrency/internal/compute/parser"
	"github.com/MussaShaukenov/go-concurrency/internal/concurrency"
	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"go.uber.org/zap"
)

type TCPServer struct {
	listener    net.Listener
	semaphore   concurrency.Semaphore
	maxConn     int
	maxMsgSize  int
	idleTimeout time.Duration
	logger      *zap.SugaredLogger
	parser      *parser.Parser
}

func NewTCPServer(logger *zap.SugaredLogger, cfg config.NetworkConfig, engine parser.Engine) (*TCPServer, error) {
	if logger == nil {
		return nil, errors.New("nil logger")
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating tcp listener: %w", err)
	}

	server := &TCPServer{
		listener:    listener,
		logger:      logger,
		maxConn:     cfg.MaxConnections,
		maxMsgSize:  cfg.MaxMessageSize,
		idleTimeout: cfg.IdleTimeout,

		parser: parser.NewParser(logger, engine),
	}

	if server.maxConn != 0 {
		server.semaphore = *concurrency.NewSemaphore(server.maxConn)
	}

	return server, nil
}

func (s *TCPServer) Start() error {
	s.logger.Infof("TCP server started on %s", s.listener.Addr().String())

	for {
		// accept connection
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Errorf("error accepting connection: %v", err)
			continue
		}

		// try to acquire a lock
		if s.maxConn != 0 {
			s.semaphore.Acquire()
		}
		// handle connection in a goroutine
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Errorf("recovered from panic: %v", err)
		}

		if s.maxConn != 0 {
			s.semaphore.Release()
		}

		if err := conn.Close(); err != nil {
			s.logger.Warn("error closing connection: %v", err)
		}
	}()

	for {
		// set deadline
		if err := conn.SetReadDeadline(time.Now().Add(s.idleTimeout)); err != nil {
			s.logger.Warn("error setting read deadline: %v", err)
		}

		// read command from client
		scanner := bufio.NewScanner(conn)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				s.logger.Warn("error scanning connection: %v", err)
				break
			}
			break
		}

		// read text command
		command := scanner.Text()

		result, err := s.parser.ParseQuery(command)
		if err != nil {
			conn.Write([]byte("ERROR: " + err.Error() + "\n"))
		}

		if result == nil && err == nil {
			conn.Write([]byte("OK\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("%v\n", result)))
		}
	}
}
