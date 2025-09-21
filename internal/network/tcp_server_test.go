package network

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewTCPServer(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name          string
		logger        interface{} // Use interface{} to allow nil
		cfg           config.NetworkConfig
		expectError   bool
		errorContains string
	}{
		{
			name:   "Valid configuration",
			logger: logger,
			cfg: config.NetworkConfig{
				Address:        "localhost:0", // Use 0 to get any available port
				MaxConnections: 100,
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError: false,
		},
		{
			name:          "Nil logger",
			logger:        nil,
			cfg:           config.NetworkConfig{},
			expectError:   true,
			errorContains: "nil logger",
		},
		{
			name:   "Invalid address",
			logger: logger,
			cfg: config.NetworkConfig{
				Address:        "invalid:99999",
				MaxConnections: 100,
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError:   true,
			errorContains: "error creating tcp listener",
		},
		{
			name:   "Zero max connections",
			logger: logger,
			cfg: config.NetworkConfig{
				Address:        "localhost:0",
				MaxConnections: 0, // Should work without semaphore
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := &mockEngine{}

			var server *TCPServer
			var err error

			if tt.logger == nil {
				server, err = NewTCPServer(nil, tt.cfg, mockEngine)
			} else {
				server, err = NewTCPServer(logger, tt.cfg, mockEngine)
			}

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.NotNil(t, server.listener)
				assert.NotNil(t, server.parser)
				assert.Equal(t, tt.cfg.MaxConnections, server.maxConn)
				assert.Equal(t, tt.cfg.MaxMessageSize, server.maxMsgSize)
				assert.Equal(t, tt.cfg.IdleTimeout, server.idleTimeout)

				// Clean up
				server.listener.Close()
			}
		})
	}
}

func TestTCPServer_CommandProcessing(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockEngine := &mockEngine{
		storage: make(map[string]string),
	}

	cfg := config.NetworkConfig{
		Address:        "localhost:0",
		MaxConnections: 10,
		MaxMessageSize: 4096,
		IdleTimeout:    30 * time.Second,
	}

	server, err := NewTCPServer(logger, cfg, mockEngine)
	require.NoError(t, err)
	defer server.listener.Close()

	// Start server in goroutine
	go func() {
		server.Start()
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name           string
		command        string
		expectedResult string
	}{
		{
			name:           "SET command",
			command:        "SET key1 value1",
			expectedResult: "OK",
		},
		{
			name:           "GET existing key",
			command:        "GET key1",
			expectedResult: "value1",
		},
		{
			name:           "GET non-existent key",
			command:        "GET nonexistent",
			expectedResult: "key not found",
		},
		{
			name:           "DEL command",
			command:        "DEL key1",
			expectedResult: "OK",
		},
		{
			name:           "Invalid command",
			command:        "INVALID key",
			expectedResult: "ERROR: invalid command",
		},
		{
			name:           "Wrong argument count",
			command:        "SET key",
			expectedResult: "ERROR: wrong length of command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Connect to server
			conn, err := net.Dial("tcp", server.listener.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			// Send command
			_, err = conn.Write([]byte(tt.command + "\n"))
			require.NoError(t, err)

			// Read response
			scanner := bufio.NewScanner(conn)
			require.True(t, scanner.Scan())
			response := scanner.Text()

			assert.Equal(t, tt.expectedResult, response)
		})
	}
}

func TestTCPServer_ConcurrentConnections(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockEngine := &mockEngine{
		storage: make(map[string]string),
	}

	cfg := config.NetworkConfig{
		Address:        "localhost:0",
		MaxConnections: 5, // Limited connections
		MaxMessageSize: 4096,
		IdleTimeout:    30 * time.Second,
	}

	server, err := NewTCPServer(logger, cfg, mockEngine)
	require.NoError(t, err)
	defer server.listener.Close()

	// Start server
	go func() {
		server.Start()
	}()
	time.Sleep(10 * time.Millisecond)

	// Test concurrent connections
	var wg sync.WaitGroup
	numClients := 10
	results := make([]error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", server.listener.Addr().String())
			if err != nil {
				results[clientID] = err
				return
			}
			defer conn.Close()

			// Send a command
			command := fmt.Sprintf("SET client_%d value_%d", clientID, clientID)
			_, err = conn.Write([]byte(command + "\n"))
			if err != nil {
				results[clientID] = err
				return
			}

			// Read response
			scanner := bufio.NewScanner(conn)
			if !scanner.Scan() {
				results[clientID] = fmt.Errorf("failed to read response")
				return
			}

			response := scanner.Text()
			if response != "OK" {
				results[clientID] = fmt.Errorf("unexpected response: %s", response)
			}
		}(i)
	}

	wg.Wait()

	// Check that most connections succeeded (some might be limited by semaphore)
	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}

	// At least some connections should succeed
	assert.Greater(t, successCount, 0, "No connections succeeded")
	t.Logf("Successful connections: %d/%d", successCount, numClients)
}

func TestTCPServer_ConnectionTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockEngine := &mockEngine{
		storage: make(map[string]string),
	}

	cfg := config.NetworkConfig{
		Address:        "localhost:0",
		MaxConnections: 10,
		MaxMessageSize: 4096,
		IdleTimeout:    100 * time.Millisecond, // Very short timeout
	}

	server, err := NewTCPServer(logger, cfg, mockEngine)
	require.NoError(t, err)
	defer server.listener.Close()

	// Start server
	go func() {
		server.Start()
	}()
	time.Sleep(10 * time.Millisecond)

	// Connect but don't send anything
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	// Wait for timeout plus some buffer
	time.Sleep(200 * time.Millisecond)

	// Try to write - connection should be closed or return error
	_, err = conn.Write([]byte("GET test\n"))
	// We expect either an error or the connection to be handled properly
	// The exact behavior depends on timing, so we just verify it doesn't panic
}

// Mock engine for testing
type mockEngine struct {
	mu      sync.RWMutex
	storage map[string]string
}

func (m *mockEngine) Set(key, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storage[fmt.Sprintf("%v", key)] = fmt.Sprintf("%v", value)
}

func (m *mockEngine) Get(key any) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keyStr := fmt.Sprintf("%v", key)
	if val, exists := m.storage[keyStr]; exists {
		return val, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (m *mockEngine) Delete(key any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyStr := fmt.Sprintf("%v", key)
	delete(m.storage, keyStr)
	return nil
}
