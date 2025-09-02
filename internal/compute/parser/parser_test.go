package parser

import (
	"errors"
	"go.uber.org/zap"
	"testing"
)

// Mock engine for testing parser independently
type mockEngine struct {
	storage  map[any]any
	setError bool
	getError bool
	delError bool
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		storage: make(map[any]any),
	}
}

func (m *mockEngine) Set(key, value any) {
	if m.setError {
		return
	}
	m.storage[key] = value
}

func (m *mockEngine) Get(key any) (any, error) {
	if m.getError {
		return nil, errors.New("mock get error")
	}
	value, exists := m.storage[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *mockEngine) Delete(key any) error {
	if m.delError {
		return errors.New("mock delete error")
	}
	if _, exists := m.storage[key]; !exists {
		return errors.New("key not found")
	}
	delete(m.storage, key)
	return nil
}

func setupTestParser() (*Parser, *mockEngine) {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	engine := newMockEngine()
	parser := NewParser(log, engine)
	return parser, engine
}

func TestParser_ParseQuery_SET(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedError error
		expectedKey   any
		expectedValue any
	}{
		{
			name:          "valid SET command",
			query:         "SET key1 value1",
			expectedError: nil,
			expectedKey:   "key1",
			expectedValue: "value1",
		},
		{
			name:          "valid SET with underscores",
			query:         "SET weather_2_pm cold_moscow_weather",
			expectedError: nil,
			expectedKey:   "weather_2_pm",
			expectedValue: "cold_moscow_weather",
		},
		{
			name:          "valid SET with slashes",
			query:         "SET /etc/nginx/config server_config",
			expectedError: nil,
			expectedKey:   "/etc/nginx/config",
			expectedValue: "server_config",
		},
		{
			name:          "valid SET with asterisks",
			query:         "SET user_**** john_doe",
			expectedError: nil,
			expectedKey:   "user_****",
			expectedValue: "john_doe",
		},
		{
			name:          "lowercase SET command",
			query:         "set key1 value1",
			expectedError: ErrInvalidCommand,
			expectedKey:   "key1",
			expectedValue: "value1",
		},
		{
			name:          "SET missing value",
			query:         "SET key1",
			expectedError: ErrWrongLength,
		},
		{
			name:          "SET too many arguments",
			query:         "SET key1 value1 extra",
			expectedError: ErrWrongLength,
		},
		{
			name:          "SET with invalid key character",
			query:         "SET key@invalid value1",
			expectedError: ErrInvalidArgument,
		},
		{
			name:          "SET with invalid value character",
			query:         "SET key1 value@invalid",
			expectedError: ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, engine := setupTestParser()

			result, err := parser.ParseQuery(tt.query)

			// Check error
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// If no error expected, check if value was set correctly
			if tt.expectedError == nil {
				storedValue, exists := engine.storage[tt.expectedKey]
				if !exists {
					t.Errorf("Expected key %v to be set in storage", tt.expectedKey)
				}
				if storedValue != tt.expectedValue {
					t.Errorf("Expected value %v, got %v", tt.expectedValue, storedValue)
				}
				if result != nil {
					t.Errorf("Expected nil result for SET command, got %v", result)
				}
			}
		})
	}
}

func TestParser_ParseQuery_GET(t *testing.T) {
	tests := []struct {
		name           string
		setupData      map[any]any
		query          string
		expectedError  error
		expectedResult any
	}{
		{
			name: "valid GET existing key",
			setupData: map[any]any{
				"key1": "value1",
			},
			query:          "GET key1",
			expectedError:  nil,
			expectedResult: "value1",
		},
		{
			name: "valid GET with underscores",
			setupData: map[any]any{
				"weather_2_pm": "cold_moscow_weather",
			},
			query:          "GET weather_2_pm",
			expectedError:  nil,
			expectedResult: "cold_moscow_weather",
		},
		{
			name: "valid GET with slashes",
			setupData: map[any]any{
				"/etc/nginx/config": "server_config",
			},
			query:          "GET /etc/nginx/config",
			expectedError:  nil,
			expectedResult: "server_config",
		},
		{
			name:           "GET non-existent key",
			setupData:      map[any]any{},
			query:          "GET nonexistent",
			expectedError:  nil,
			expectedResult: "key not found",
		},
		{
			name:           "lowercase GET command",
			setupData:      map[any]any{"key1": "value1"},
			query:          "get key1",
			expectedError:  ErrInvalidCommand,
			expectedResult: nil,
		},
		{
			name:          "GET no arguments",
			query:         "GET",
			expectedError: ErrWrongLength,
		},
		{
			name:          "GET too many arguments",
			query:         "GET key1 extra",
			expectedError: ErrWrongLength,
		},
		{
			name:          "GET with invalid key character",
			query:         "GET key@invalid",
			expectedError: ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, engine := setupTestParser()

			// Setup test data
			for k, v := range tt.setupData {
				engine.storage[k] = v
			}

			result, err := parser.ParseQuery(tt.query)

			// Check error
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Check result
			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestParser_ParseQuery_DEL(t *testing.T) {
	tests := []struct {
		name           string
		setupData      map[any]any
		query          string
		expectedError  error
		expectedResult any
		keyDeleted     any
	}{
		{
			name: "valid DEL existing key",
			setupData: map[any]any{
				"key1": "value1",
				"key2": "value2",
			},
			query:          "DEL key1",
			expectedError:  nil,
			expectedResult: nil,
			keyDeleted:     "key1",
		},
		{
			name: "valid DEL with underscores",
			setupData: map[any]any{
				"user_****": "john_doe",
			},
			query:          "DEL user_****",
			expectedError:  nil,
			expectedResult: nil,
			keyDeleted:     "user_****",
		},
		{
			name:           "DEL non-existent key",
			setupData:      map[any]any{},
			query:          "DEL nonexistent",
			expectedError:  nil,
			expectedResult: "key not found",
		},
		{
			name: "lowercase DEL command",
			setupData: map[any]any{
				"key1": "value1",
			},
			query:          "del key1",
			expectedError:  ErrInvalidCommand,
			expectedResult: nil,
			keyDeleted:     "key1",
		},
		{
			name:          "DEL no arguments",
			query:         "DEL",
			expectedError: ErrWrongLength,
		},
		{
			name:          "DEL too many arguments",
			query:         "DEL key1 extra",
			expectedError: ErrWrongLength,
		},
		{
			name:          "DEL with invalid key character",
			query:         "DEL key@invalid",
			expectedError: ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, engine := setupTestParser()

			// Setup test data
			for k, v := range tt.setupData {
				engine.storage[k] = v
			}

			result, err := parser.ParseQuery(tt.query)

			// Check error
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Check result
			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			// Check if key was deleted (for successful deletions)
			if tt.expectedError == nil && tt.expectedResult == nil {
				if _, exists := engine.storage[tt.keyDeleted]; exists {
					t.Errorf("Expected key %v to be deleted", tt.keyDeleted)
				}
			}
		})
	}
}

func TestParser_ParseQuery_InvalidCommands(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedError error
	}{
		{
			name:          "invalid command",
			query:         "INVALID key1",
			expectedError: ErrInvalidCommand,
		},
		{
			name:          "empty query",
			query:         "",
			expectedError: ErrInvalidCommand,
		},
		{
			name:          "only spaces",
			query:         "   ",
			expectedError: ErrInvalidCommand,
		},
		{
			name:          "unknown command",
			query:         "UPDATE key1 value1",
			expectedError: ErrInvalidCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, _ := setupTestParser()

			_, err := parser.ParseQuery(tt.query)

			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestParser_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedError error
		description   string
	}{
		{
			name:          "valid characters - letters and digits",
			query:         "SET key123 value456",
			expectedError: nil,
			description:   "letters and digits should be allowed",
		},
		{
			name:          "valid characters - underscores",
			query:         "SET key_name value_data",
			expectedError: nil,
			description:   "underscores should be allowed",
		},
		{
			name:          "valid characters - asterisks",
			query:         "SET key**** value****",
			expectedError: nil,
			description:   "asterisks should be allowed",
		},
		{
			name:          "valid characters - forward slashes",
			query:         "SET /path/to/key /path/to/value",
			expectedError: nil,
			description:   "forward slashes should be allowed",
		},
		{
			name:          "invalid character - at symbol",
			query:         "SET key@symbol value",
			expectedError: ErrInvalidArgument,
			description:   "at symbol should not be allowed",
		},
		{
			name:          "invalid character - hash",
			query:         "SET key# value",
			expectedError: ErrInvalidArgument,
			description:   "hash should not be allowed",
		},
		{
			name:          "invalid character - percent",
			query:         "SET key value%",
			expectedError: ErrInvalidArgument,
			description:   "percent should not be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, _ := setupTestParser()

			_, err := parser.ParseQuery(tt.query)

			if err != tt.expectedError {
				t.Errorf("%s: Expected error %v, got %v", tt.description, tt.expectedError, err)
			}
		})
	}
}
