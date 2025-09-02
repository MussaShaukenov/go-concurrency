package engine

import (
	"go.uber.org/zap"
	"testing"
)

func setupTestEngine() *Engine {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	return NewEngine(log)
}

func TestEngine_Set(t *testing.T) {
	tests := []struct {
		name  string
		key   any
		value any
	}{
		{
			name:  "string key and value",
			key:   "key1",
			value: "value1",
		},
		{
			name:  "integer key and value",
			key:   123,
			value: 456,
		},
		{
			name:  "mixed types",
			key:   "key2",
			value: 789,
		},
		{
			name:  "empty string key",
			key:   "",
			value: "empty_key_value",
		},
		{
			name:  "nil value",
			key:   "nil_key",
			value: nil,
		},
		{
			name:  "overwrite existing key",
			key:   "existing_key",
			value: "new_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := setupTestEngine()

			// For overwrite test, set initial value
			if tt.name == "overwrite existing key" {
				engine.Set(tt.key, "old_value")
			}

			engine.Set(tt.key, tt.value)

			// Verify the value was set
			storedValue, exists := engine.storage[tt.key]
			if !exists {
				t.Errorf("Expected key %v to exist in storage", tt.key)
			}
			if storedValue != tt.value {
				t.Errorf("Expected value %v, got %v", tt.value, storedValue)
			}
		})
	}
}

func TestEngine_Get(t *testing.T) {
	tests := []struct {
		name          string
		setupData     map[any]any
		key           any
		expectedValue any
		expectedError error
	}{
		{
			name: "get existing string key",
			setupData: map[any]any{
				"key1": "value1",
			},
			key:           "key1",
			expectedValue: "value1",
			expectedError: nil,
		},
		{
			name: "get existing integer key",
			setupData: map[any]any{
				123: "int_key_value",
			},
			key:           123,
			expectedValue: "int_key_value",
			expectedError: nil,
		},
		{
			name:          "get non-existent key",
			setupData:     map[any]any{},
			key:           "nonexistent",
			expectedValue: nil,
			expectedError: ErrKeyNotFound,
		},
		{
			name: "get key with nil value",
			setupData: map[any]any{
				"nil_key": nil,
			},
			key:           "nil_key",
			expectedValue: nil,
			expectedError: nil,
		},
		{
			name: "get empty string key",
			setupData: map[any]any{
				"": "empty_key_value",
			},
			key:           "",
			expectedValue: "empty_key_value",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := setupTestEngine()

			// Setup test data
			for k, v := range tt.setupData {
				engine.storage[k] = v
			}

			value, err := engine.Get(tt.key)

			// Check error
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Check value
			if value != tt.expectedValue {
				t.Errorf("Expected value %v, got %v", tt.expectedValue, value)
			}
		})
	}
}

func TestEngine_Delete(t *testing.T) {
	tests := []struct {
		name          string
		setupData     map[any]any
		keyToDelete   any
		expectedError error
		shouldExist   bool
	}{
		{
			name: "delete existing string key",
			setupData: map[any]any{
				"key1": "value1",
				"key2": "value2",
			},
			keyToDelete:   "key1",
			expectedError: nil,
			shouldExist:   false,
		},
		{
			name: "delete existing integer key",
			setupData: map[any]any{
				123:    "int_value",
				"key1": "value1",
			},
			keyToDelete:   123,
			expectedError: nil,
			shouldExist:   false,
		},
		{
			name:          "delete non-existent key",
			setupData:     map[any]any{},
			keyToDelete:   "nonexistent",
			expectedError: ErrKeyNotFound,
			shouldExist:   false,
		},
		{
			name: "delete key with nil value",
			setupData: map[any]any{
				"nil_key": nil,
			},
			keyToDelete:   "nil_key",
			expectedError: nil,
			shouldExist:   false,
		},
		{
			name: "delete empty string key",
			setupData: map[any]any{
				"": "empty_key_value",
			},
			keyToDelete:   "",
			expectedError: nil,
			shouldExist:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := setupTestEngine()
			initialSize := len(tt.setupData)

			// Setup test data
			for k, v := range tt.setupData {
				engine.storage[k] = v
			}

			err := engine.Delete(tt.keyToDelete)

			// Check error
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}

			// Check if key exists after deletion
			_, exists := engine.storage[tt.keyToDelete]
			if exists != tt.shouldExist {
				t.Errorf("Expected key existence %v, got %v", tt.shouldExist, exists)
			}

			// Check storage size after successful deletion
			if tt.expectedError == nil {
				expectedSize := initialSize - 1
				if len(engine.storage) != expectedSize {
					t.Errorf("Expected storage size %d, got %d", expectedSize, len(engine.storage))
				}
			} else {
				// Size should remain unchanged for failed deletions
				if len(engine.storage) != initialSize {
					t.Errorf("Expected storage size %d after failed deletion, got %d", initialSize, len(engine.storage))
				}
			}
		})
	}
}

func TestEngine_Operations_Integration(t *testing.T) {
	tests := []struct {
		name       string
		operations []struct {
			op    string
			key   any
			value any
		}
		finalState map[any]any
	}{
		{
			name: "set, get, delete sequence",
			operations: []struct {
				op    string
				key   any
				value any
			}{
				{"set", "key1", "value1"},
				{"set", "key2", "value2"},
				{"delete", "key1", nil},
			},
			finalState: map[any]any{
				"key2": "value2",
			},
		},
		{
			name: "overwrite and delete",
			operations: []struct {
				op    string
				key   any
				value any
			}{
				{"set", "key1", "original"},
				{"set", "key1", "updated"},
				{"delete", "key1", nil},
			},
			finalState: map[any]any{},
		},
		{
			name: "mixed key types",
			operations: []struct {
				op    string
				key   any
				value any
			}{
				{"set", "string_key", "string_value"},
				{"set", 123, "int_key_value"},
				{"set", "", "empty_key_value"},
			},
			finalState: map[any]any{
				"string_key": "string_value",
				123:          "int_key_value",
				"":           "empty_key_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := setupTestEngine()

			// Execute operations
			for _, op := range tt.operations {
				switch op.op {
				case "set":
					engine.Set(op.key, op.value)
				case "delete":
					engine.Delete(op.key)
				}
			}

			// Verify final state
			if len(engine.storage) != len(tt.finalState) {
				t.Errorf("Expected storage size %d, got %d", len(tt.finalState), len(engine.storage))
			}

			for expectedKey, expectedValue := range tt.finalState {
				actualValue, exists := engine.storage[expectedKey]
				if !exists {
					t.Errorf("Expected key %v to exist in final state", expectedKey)
				}
				if actualValue != expectedValue {
					t.Errorf("For key %v, expected value %v, got %v", expectedKey, expectedValue, actualValue)
				}
			}
		})
	}
}
