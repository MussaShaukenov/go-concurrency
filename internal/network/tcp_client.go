package network

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
)

type TCPClient struct {
	conn           net.Conn
	idleTimeout    time.Duration
	maxMessageSize int
}

func NewTCPClient(cfg config.NetworkConfig) (*TCPClient, error) {
	connection, err := net.Dial("tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}

	client := &TCPClient{
		conn:           connection,
		maxMessageSize: cfg.MaxMessageSize,
		idleTimeout:    cfg.IdleTimeout,
	}

	if client.idleTimeout != 0 {
		if err := client.conn.SetDeadline(time.Now().Add(client.idleTimeout)); err != nil {
			return nil, fmt.Errorf("error setting idle timeout: %w", err)
		}
	}

	return client, nil
}

func (c *TCPClient) SendCommand(cmd string) (string, error) {
	// send command
	_, err := c.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", fmt.Errorf("error sending command: %w", err)
	}

	scanner := bufio.NewScanner(c.conn)
	if !scanner.Scan() {
		return "", fmt.Errorf("error reading response: %w", scanner.Err())
	}

	return scanner.Text(), scanner.Err()
}

func (c *TCPClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
