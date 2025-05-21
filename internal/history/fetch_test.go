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

func TestSaveFullHistoryAndUpdateNickname(t *testing.T) {
	tempDir := t.TempDir()
	tempHistoryFilePath := filepath.Join(tempDir, "test_history_save_fetch.json")

	// Override getFileLocation for this test
	originalGetFileLocation := getFileLocation
	getFileLocation = func() string {
		return tempHistoryFilePath
	}
	defer func() { getFileLocation = originalGetFileLocation }()

	// 1. Initial Save and Fetch
	now := time.Now().UTC().Truncate(time.Second) // Ensure consistent, truncated time
	entry1 := SSHHistory{Connection: config.SSHConfig{Name: "host1", Nickname: "server-alpha", Host: "1.1.1.1", User: "user1", Port: "22", Key: "key1"}, Date: now}
	entry2 := SSHHistory{Connection: config.SSHConfig{Name: "host2", Nickname: "", Host: "2.2.2.2", User: "user2", Port: "22", Key: "key2"}, Date: now.Add(-time.Hour)}        // Different date
	entry3 := SSHHistory{Connection: config.SSHConfig{Name: "host3", Nickname: "server-beta", Host: "3.3.3.3", User: "user3", Port: "22", Key: "key3"}, Date: now.Add(-2 * time.Hour)} // Different date
	initialList := []SSHHistory{entry1, entry2, entry3}

	// Ensure all Date fields in initialList are truncated for fair comparison with fetched data
	for i := range initialList {
		initialList[i].Date = initialList[i].Date.Truncate(time.Second)
	}

	err := SaveFullHistory(initialList)
	if err != nil {
		t.Fatalf("SaveFullHistory (initial) failed: %v", err)
	}

	fileContent, err := os.ReadFile(tempHistoryFilePath)
	if err != nil {
		t.Fatalf("Failed to read temp history file: %v", err)
	}
	fetchedList1, err := Fetch(fileContent)
	if err != nil {
		t.Fatalf("Fetch (initial) failed: %v", err)
	}

	// Truncate dates in fetchedList1 before comparison
	for i := range fetchedList1 {
		fetchedList1[i].Date = fetchedList1[i].Date.Truncate(time.Second)
	}

	if !reflect.DeepEqual(initialList, fetchedList1) {
		// For more detailed diff in case of mismatch:
		for i := 0; i < len(initialList); i++ {
			if i >= len(fetchedList1) {
				t.Errorf("Fetched list is shorter. Missing entry at index %d: %v", i, initialList[i])
				break
			}
			if !reflect.DeepEqual(initialList[i], fetchedList1[i]) {
				t.Errorf("Entry mismatch at index %d:\nOriginal: %+v\nFetched:  %+v", i, initialList[i], fetchedList1[i])
			}
		}
		if len(fetchedList1) > len(initialList) {
			t.Errorf("Fetched list is longer. Extra entry at index %d: %v", len(initialList), fetchedList1[len(initialList)])
		}
		t.Fatalf("Initial fetched list does not match original list. See details above.")
	}

	// 2. Update Nickname, Save, and Fetch Again
	updatedList := make([]SSHHistory, len(fetchedList1))
	copy(updatedList, fetchedList1) // Start with a fresh copy of the correctly fetched list

	// Apply updates (ensure Date fields remain truncated from previous step)
	updatedList[0].Connection.Nickname = "server-gamma"       // Update nickname
	updatedList[1].Connection.Nickname = "new-nickname-for-2" // Add nickname
	updatedList[2].Connection.Nickname = ""                   // Update nickname of the third entry to empty

	err = SaveFullHistory(updatedList)
	if err != nil {
		t.Fatalf("SaveFullHistory (updated) failed: %v", err)
	}

	fileContent2, err := os.ReadFile(tempHistoryFilePath)
	if err != nil {
		t.Fatalf("Failed to read temp history file (after update): %v", err)
	}
	fetchedList2, err := Fetch(fileContent2)
	if err != nil {
		t.Fatalf("Fetch (updated) failed: %v", err)
	}

	// Truncate dates in fetchedList2 before comparison
	for i := range fetchedList2 {
		fetchedList2[i].Date = fetchedList2[i].Date.Truncate(time.Second)
	}
	
	// Before DeepEqual, ensure updatedList also has its Date fields appropriately set (should be fine from copy if fetchedList1 was truncated)
	if !reflect.DeepEqual(updatedList, fetchedList2) {
		for i := 0; i < len(updatedList); i++ {
			if i >= len(fetchedList2) {
				t.Errorf("Fetched list2 is shorter. Missing entry at index %d: %v", i, updatedList[i])
				break
			}
			if !reflect.DeepEqual(updatedList[i], fetchedList2[i]) {
				t.Errorf("Entry mismatch at index %d (update stage):\nExpected: %+v\nFetched:  %+v", i, updatedList[i], fetchedList2[i])
			}
		}
		if len(fetchedList2) > len(updatedList) {
			t.Errorf("Fetched list2 is longer. Extra entry at index %d: %v", len(updatedList), fetchedList2[len(updatedList)])
		}
		t.Fatalf("Updated fetched list does not match expected updated list. See details above.")
	}

	// Verify specific changes
	if fetchedList2[0].Connection.Nickname != "server-gamma" {
		t.Errorf("Expected nickname 'server-gamma' for first entry, got '%s'", fetchedList2[0].Connection.Nickname)
	}
	if fetchedList2[1].Connection.Nickname != "new-nickname-for-2" {
		t.Errorf("Expected nickname 'new-nickname-for-2' for second entry, got '%s'", fetchedList2[1].Connection.Nickname)
	}
	if fetchedList2[2].Connection.Nickname != "" { // Verify it's empty
		t.Errorf("Expected nickname '' (empty string) for third entry, got '%s'", fetchedList2[2].Connection.Nickname)
	}
}
