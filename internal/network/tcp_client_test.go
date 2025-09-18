package network

import (
	"net"
	"testing"
	"time"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPClient_SendCommand(t *testing.T) {
	// Start a mock server for testing
	mockServer := startMockServer(t)
	defer mockServer.Close()

	tests := []struct {
		name           string
		command        string
		serverResponse string
		expectedResult string
		expectedError  bool
	}{
		{
			name:           "SET command success",
			command:        "SET key value",
			serverResponse: "OK\n",
			expectedResult: "OK",
			expectedError:  false,
		},
		{
			name:           "GET command success",
			command:        "GET key",
			serverResponse: "value\n",
			expectedResult: "value",
			expectedError:  false,
		},
		{
			name:           "DEL command success",
			command:        "DEL key",
			serverResponse: "OK\n",
			expectedResult: "OK",
			expectedError:  false,
		},
		{
			name:           "ERROR response",
			command:        "INVALID command",
			serverResponse: "ERROR: invalid command\n",
			expectedResult: "ERROR: invalid command",
			expectedError:  false,
		},
		{
			name:           "Empty command",
			command:        "",
			serverResponse: "ERROR: empty command\n",
			expectedResult: "ERROR: empty command",
			expectedError:  false,
		},
		{
			name:           "Key not found",
			command:        "GET nonexistent",
			serverResponse: "key not found\n",
			expectedResult: "key not found",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure the mock server to respond with expected response
			mockServer.setResponse(tt.serverResponse)

			// Create client
			cfg := config.NetworkConfig{
				Address:        mockServer.Address(),
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			}

			client, err := NewTCPClient(cfg)
			require.NoError(t, err, "Failed to create TCP client")
			defer client.Close()

			// Send command and check result
			result, err := client.SendCommand(tt.command)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestNewTCPClient(t *testing.T) {
	tests := []struct {
		name          string
		config        config.NetworkConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid config",
			config: config.NetworkConfig{
				Address:        "localhost:0", // Will be replaced with actual mock server address
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "Invalid address",
			config: config.NetworkConfig{
				Address:        "invalid:99999",
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError:   true,
			errorContains: "error connecting to server",
		},
		{
			name: "Non-existent host",
			config: config.NetworkConfig{
				Address:        "non-existent-host:8080",
				MaxMessageSize: 4096,
				IdleTimeout:    30 * time.Second,
			},
			expectError:   true,
			errorContains: "error connecting to server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Valid config" {
				// Start mock server for valid config test
				mockServer := startMockServer(t)
				defer mockServer.Close()
				tt.config.Address = mockServer.Address()
			}

			client, err := NewTCPClient(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.conn)
				assert.Equal(t, tt.config.MaxMessageSize, client.maxMessageSize)
				assert.Equal(t, tt.config.IdleTimeout, client.idleTimeout)

				client.Close()
			}
		})
	}
}

func TestTCPClient_Close(t *testing.T) {
	mockServer := startMockServer(t)
	defer mockServer.Close()

	cfg := config.NetworkConfig{
		Address:        mockServer.Address(),
		MaxMessageSize: 4096,
		IdleTimeout:    30 * time.Second,
	}

	client, err := NewTCPClient(cfg)
	require.NoError(t, err)

	// Close should not panic and should be safe to call multiple times
	client.Close()
	client.Close() // Should not panic

	// After close, SendCommand should fail
	_, err = client.SendCommand("GET test")
	assert.Error(t, err)
}

// Mock server for testing
type mockServer struct {
	listener net.Listener
	response string
}

func startMockServer(t *testing.T) *mockServer {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := &mockServer{
		listener: listener,
		response: "OK\n", // Default response
	}

	go server.handleConnections()

	return server
}

func (ms *mockServer) Address() string {
	return ms.listener.Addr().String()
}

func (ms *mockServer) setResponse(response string) {
	ms.response = response
}

func (ms *mockServer) Close() {
	ms.listener.Close()
}

func (ms *mockServer) handleConnections() {
	for {
		conn, err := ms.listener.Accept()
		if err != nil {
			return // Server closed
		}

		go ms.handleConnection(conn)
	}
}

func (ms *mockServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the command (we don't really need to parse it for testing)
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		return
	}

	// Send the configured response
	conn.Write([]byte(ms.response))
}

func TestTCPClient_SendCommand_ConnectionErrors(t *testing.T) {
	// Test connection error during NewTCPClient
	t.Run("Connection refused", func(t *testing.T) {
		cfg := config.NetworkConfig{
			Address:        "127.0.0.1:99999", // Port that should be closed
			MaxMessageSize: 4096,
			IdleTimeout:    30 * time.Second,
		}

		client, err := NewTCPClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "error connecting to server")
	})

	// Test closed connection after successful creation
	t.Run("Connection closed after creation", func(t *testing.T) {
		mockServer := startMockServer(t)

		cfg := config.NetworkConfig{
			Address:        mockServer.Address(),
			MaxMessageSize: 4096,
			IdleTimeout:    30 * time.Second,
		}

		client, err := NewTCPClient(cfg)
		require.NoError(t, err)

		// Close the client's connection directly
		client.conn.Close()

		// This should fail because connection is closed
		_, err = client.SendCommand("GET test")
		assert.Error(t, err)

		mockServer.Close()
		client.Close() // Safe to call even if already closed
	})
}
