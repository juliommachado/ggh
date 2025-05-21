package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/byawitz/ggh/internal/config"
)

var historyFile = `
[
  {
    "time": "2024-01-01 00:00:00 +0000 UTC",
    "connection": {
      "name": "stage",
      "host": "host.name",
      "key": "~/.ssh/id_rsa"
    }
  },
  {
    "time": "2022-01-01 00:00:00 +0000 UTC",
    "connection": {
      "name": "production",
      "host": "host2.name",
      "port": "5412",
      "user": "ubuntu"
    }
  }
]
`

func TestParsing(t *testing.T) {
	history, err := Fetch([]byte(historyFile))

	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Parsing config file failed: got %v, want %v\n", len(history), 2) // Corrected expected length
	}

	if history[0].Connection.Host != "host.name" {
		t.Errorf("Parsing config file failed: got %v, want %v\n", history[0].Connection.Host, "host.name")
	}

	if history[0].Connection.Port != "" {
		t.Errorf("Parsing config file failed: got %v, want %v\n", history[0].Connection.Port, "")
	}

	if history[0].Connection.User != "" {
		t.Errorf("Parsing config file failed: got %v, want %v\n", history[0].Connection.User, "") // Corrected field from Port to User
	}

	if history[1].Connection.Host != "host2.name" {
		t.Errorf("Parsing config file failed: got %v, want %v\n", history[1].Connection.Host, "host2.name") // Corrected expected host
	}
}

func TestFetchWithNickname(t *testing.T) {
	// 1. Setup
	originalConfig := config.SSHConfig{
		Name:     "test-host-with-nickname",
		Nickname: "my-special-server",
		Host:     "1.2.3.4",
		User:     "testuser",
		Port:     "22",
		Key:      "~/.ssh/id_test",
	}
	originalHistoryEntry := SSHHistory{
		Connection: originalConfig,
		Date:       time.Now().UTC(), // Use UTC for consistency
	}
	historyList := []SSHHistory{originalHistoryEntry}

	// Create a temporary directory for the test file
	tempDir := t.TempDir() // t.TempDir() handles cleanup automatically
	tempHistoryFile := filepath.Join(tempDir, "history_with_nickname.json")

	// 2. Action: Marshal and write to temp file
	jsonData, err := json.MarshalIndent(historyList, "", "  ") // Use MarshalIndent for readability if needed
	if err != nil {
		t.Fatalf("Failed to marshal test history data: %v", err)
	}
	err = os.WriteFile(tempHistoryFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write temporary history file: %v", err)
	}

	// Read the file content for Fetch
	fileContent, err := os.ReadFile(tempHistoryFile)
	if err != nil {
		t.Fatalf("Failed to read back temporary history file: %v", err)
	}

	// Fetch the history
	fetchedList, err := Fetch(fileContent)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	// 3. Assertion
	if len(fetchedList) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(fetchedList))
	}

	fetchedEntry := fetchedList[0]
	fetchedConfig := fetchedEntry.Connection

	if fetchedConfig.Nickname != originalConfig.Nickname {
		t.Errorf("Expected Nickname '%s', got '%s'", originalConfig.Nickname, fetchedConfig.Nickname)
	}
	if fetchedConfig.Name != originalConfig.Name {
		t.Errorf("Expected Name '%s', got '%s'", originalConfig.Name, fetchedConfig.Name)
	}
	if fetchedConfig.Host != originalConfig.Host {
		t.Errorf("Expected Host '%s', got '%s'", originalConfig.Host, fetchedConfig.Host)
	}
	if fetchedConfig.User != originalConfig.User {
		t.Errorf("Expected User '%s', got '%s'", originalConfig.User, fetchedConfig.User)
	}
	if fetchedConfig.Port != originalConfig.Port {
		t.Errorf("Expected Port '%s', got '%s'", originalConfig.Port, fetchedConfig.Port)
	}
	if fetchedConfig.Key != originalConfig.Key {
		t.Errorf("Expected Key '%s', got '%s'", originalConfig.Key, fetchedConfig.Key)
	}

	// Compare dates by truncating to second to avoid issues with monotonic clock readings or sub-second precision differences
	// after marshalling/unmarshalling.
	if !fetchedEntry.Date.Truncate(time.Second).Equal(originalHistoryEntry.Date.Truncate(time.Second)) {
		t.Errorf("Expected Date '%v', got '%v'", originalHistoryEntry.Date, fetchedEntry.Date)
	}
}
