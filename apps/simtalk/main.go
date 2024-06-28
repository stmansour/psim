package main

import (
	"bufio"
	"encoding/json"
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

// SimulatorStatus represents the status information of the simulator
type SimulatorStatus struct {
	ProgramStarted         string `json:"ProgramStarted"`
	RunDuration            string `json:"RunDuration"`
	ConfigFile             string `json:"ConfigFile"`
	SimulationDateRange    string `json:"SimulationDateRange"`
	PopulationSize         int    `json:"PopulationSize"`
	LoopCount              int    `json:"LoopCount"`
	GenerationsRequested   int    `json:"GenerationsRequested"`
	CompletedLoops         int    `json:"CompletedLoops"`
	CompletedGenerations   int    `json:"CompletedGenerations"`
	ElapsedTimeLastGen     string `json:"ElapsedTimeLastGen"`
	EstimatedTimeRemaining string `json:"EstimatedTimeRemaining"`
	EstimatedCompletion    string `json:"EstimatedCompletion"`
}

// StopResponse represents the response from the stop command
type StopResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

func main() {
	port := 8090
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

	switch command {
	case "status":
		var status SimulatorStatus
		err = json.Unmarshal(body, &status)
		if err != nil {
			return "", fmt.Errorf("error unmarshaling response body: %v", err)
		}
		return formatSimulatorStatus(status), nil
	case "stop":
		var stopResp StopResponse
		err = json.Unmarshal(body, &stopResp)
		if err != nil {
			return "", fmt.Errorf("error unmarshaling response body: %v", err)
		}
		return fmt.Sprintf("Status: %s\nMessage: %s\n", stopResp.Status, stopResp.Message), nil
	default:
		// Convert the response body to a string for non-status and non-stop commands
		return string(body), nil
	}
}

func formatSimulatorStatus(status SimulatorStatus) string {
	return fmt.Sprintf(
		"SIMULATOR STATUS\n"+
			"                    Program started: %s\n"+
			"                Run duration so far: %s\n"+
			"                        Config file: %s\n"+
			"              Simulation Date Range: %s\n"+
			"                    Population size: %d\n"+
			"LoopCount and Generations requested: %d loops, %d generations\n"+
			"                          completed: %d loops, %d generations\n"+
			"       Elapsed time last generation: %s\n"+
			"           Estimated time remaining: %s\n"+
			"               Estimated completion: %s\n",
		status.ProgramStarted,
		status.RunDuration,
		status.ConfigFile,
		status.SimulationDateRange,
		status.PopulationSize,
		status.LoopCount,
		status.GenerationsRequested,
		status.CompletedLoops,
		status.CompletedGenerations,
		status.ElapsedTimeLastGen,
		status.EstimatedTimeRemaining,
		status.EstimatedCompletion,
	)
}
