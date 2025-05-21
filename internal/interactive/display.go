package interactive

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/byawitz/ggh/internal/config"
	"github.com/byawitz/ggh/internal/history"
	"github.com/byawitz/ggh/internal/ssh"
	"github.com/byawitz/ggh/internal/theme"
	"github.com/charmbracelet/bubbles/table"
)

func Config() []string {
	list, err := config.Parse(config.GetConfigFile())

	if err != nil {
		log.Fatal(err)
	}

	if len(list) == 0 {
		fmt.Println("No configs found in ~/.ssh/config.")
		os.Exit(0)
	}

	var rows []table.Row
	for _, configItem := range list {
		rows = append(rows, table.Row{configItem.Name, configItem.Host, configItem.Port, configItem.User, configItem.Key})
	}
	selectedConfig, _, _ := Select(rows, SelectConfig)

	if selectedConfig.Host == "" {
		os.Exit(0)
	}

	history.Add(selectedConfig)

	return ssh.GenerateCommandArgs(selectedConfig)
}

func History() []string {
	list, err := history.FetchWithDefaultFile()
	if err != nil {
		log.Fatal(err)
	}
	if len(list) == 0 {
		fmt.Println("No history found.")
		os.Exit(0)
	}

	for { // Main interaction loop
		// Generate/Refresh rows
		var rows []table.Row
		currentTime := time.Now()
		for _, hi := range list { // Use 'hi' to avoid conflict if 'historyItem' is used later
			rows = append(rows, table.Row{
				hi.Connection.Name,
				hi.Connection.Nickname,
				hi.Connection.Host,
				hi.Connection.Port,
				hi.Connection.User,
				hi.Connection.Key,
				fmt.Sprintf("%s", history.ReadableTime(currentTime.Sub(hi.Date))),
			})
		}

		selectedConfig, action, index := Select(rows, SelectHistory)

		switch action {
		case "edit":
			if index < 0 || index >= len(list) {
				log.Println("Error: Invalid selection index for edit.")
				continue
			}
			historyItemToEdit := &list[index]

			fmt.Printf("Enter new nickname for %s (Host: %s) (or press Enter to keep '%s'): ", historyItemToEdit.Connection.Name, historyItemToEdit.Connection.Host, historyItemToEdit.Connection.Nickname)
			reader := bufio.NewReader(os.Stdin)
			newNicknameInput, errReader := reader.ReadString('\n')
			if errReader != nil {
				log.Printf("Error reading nickname: %v\n", errReader)
				continue
			}
			newNicknameInput = strings.TrimSpace(newNicknameInput)

			if newNicknameInput != "" {
				historyItemToEdit.Connection.Nickname = newNicknameInput
			}

			// This function (SaveFullHistory) will be created in the next sub-task.
			// For now, the sub-task is just to ensure this call is made.
			errSave := history.SaveFullHistory(list)
			if errSave != nil {
				log.Fatalf("Error saving history: %v\n", errSave)
			}
			// Loop continues, rows are regenerated, Select is called again.
			continue
		case "connect":
			return ssh.GenerateCommandArgs(selectedConfig)
		case "quit":
			os.Exit(0)
		default:
			log.Printf("Unknown action: %s\n", action)
			os.Exit(1)
		}
	}
}
