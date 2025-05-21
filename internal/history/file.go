package history

import (
	"os"
	"path/filepath"
)

var getFileLocation = func() string { // Changed to var
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return ""
	}

	gghConfigDir := filepath.Join(userHomeDir, ".ggh")

	if err := os.MkdirAll(gghConfigDir, 0700); err != nil {
		return ""
	}

	return filepath.Join(gghConfigDir, "history.json")

}
func getFile() []byte {

	history, err := os.ReadFile(getFileLocation())

	if err != nil {
		return []byte{}
	}

	return history
}
