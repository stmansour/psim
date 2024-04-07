package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Command is a struct that holds the command name and its description
type Command struct {
	Name        string
	Description string
}

func main() {
	port := 8080
	var err error
	if len(os.Args) > 1 {
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("Invalid port number: %s\n", os.Args[1])
			os.Exit(1)
		}
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	commands := []Command{
		{"status", "Get the current status of the simulator."},
		{"stop", "Tell the simulator to stop after completing the current generation."},
	}

	resp, err := http.Get(baseURL + "/status")
	if err != nil {
		fmt.Printf("Failed to connect to the simulator on port %d. Ensure it is running.\n", port)
		os.Exit(1)
	}
	resp.Body.Close()

	fmt.Println("Connected to the simulator. Enter 'help' for a list of commands.")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("@simulator:%d > ", port)
		scanned := scanner.Scan()
		if !scanned {
			break
		}

		text := scanner.Text()
		trimmedText := strings.TrimSpace(text)

		if trimmedText == "help" {
			fmt.Println("Available commands:")
			for _, cmd := range commands {
				fmt.Printf("- %s : %s\n", cmd.Name, cmd.Description)
			}
			continue
		} else if trimmedText == "quit" || trimmedText == "exit" {
			break
		}

		response, err := sendCommand(baseURL, trimmedText)
		if err != nil {
			fmt.Printf("Error sending command: %v\n", err)
			continue
		}

		fmt.Println(response)
	}
}

func sendCommand(baseURL, command string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, command))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned error status: %s", resp.Status)
	}

	// Convert the response body to a string
	return string(body), nil
}
