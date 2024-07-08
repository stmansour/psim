package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

// Command is a struct that holds the command name and its description
type Command struct {
	Name        string
	Description string
}

// SimulatorStatus represents the status information of the simulator
type SimulatorStatus struct {
	ProgramStarted         string
	RunDuration            string
	ConfigFile             string
	SimulationDateRange    string
	PopulationSize         int
	LoopCount              int
	GenerationsRequested   int
	CompletedLoops         int
	CompletedGenerations   int
	ElapsedTimeLastGen     string
	EstimatedTimeRemaining string
	EstimatedCompletion    string
	SID                    int64
}

// StopResponse represents the response from the stop command
type StopResponse struct {
	Status  string
	Message string
}

var app struct {
	ports []int // list of ports between 8090 and 8100 with listeners
	pidx  int   // index into the list of ports
}

// checkPort tries to establish a TCP connection to the given port and returns true if successful
func checkPort(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// scanPorts scans the specified range of ports and returns a list of ports with listeners
func scanPorts(startPort, endPort int) []int {
	var openPorts []int
	for port := startPort; port <= endPort; port++ {
		if checkPort(port) {
			openPorts = append(openPorts, port)
		}
	}
	return openPorts
}

func main() {
	var port int

	//---------------------------------------------------
	// SEE IF THERE ARE ANY SIMULATOR PROCESSES RUNNING
	//---------------------------------------------------
	app.ports = scanPorts(8090, 8100)
	if len(app.ports) == 0 {
		app.pidx = -1
	}

	//---------------------------------------------------
	// SELECT THE ONE THE USER INDICATED IF POSSIBLE...
	//---------------------------------------------------
	var err error
	if len(os.Args) > 1 {
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("Invalid port number: %s\n", os.Args[1])
			os.Exit(1)
		}
		for i := 0; i < len(app.ports); i++ {
			if app.ports[i] == port {
				app.pidx = i
				break
			}
		}
		if app.pidx != port {
			fmt.Printf("No simulator processes running on port %d.\n", port)
			fmt.Printf("Currently selected simulator process is listening on port %d.\n", app.ports[app.pidx])
		}
	}

	commands := []Command{
		{"help", "List the available commands."},
		{"next", "Select the next simulator in the list."},
		{"port <portnumber>", "Switch to port <portnumber>."},
		{"ports", "List the available ports."},
		{"prev", "Select the previous simulator in the list."},
		{"rescan", "Rescan the list of available ports."},
		{"status", "Get the current status of the simulator."},
		{"stopsim", "Tell the simulator to stop after completing the current generation."},
	}

	if len(app.ports) > 1 {
		fmt.Printf("Found %d simulators running on ports these ports:\n", len(app.ports))
		for i := 0; i < len(app.ports); i++ {
			fmt.Printf("  %d\n", app.ports[i])
		}
	}
	if len(app.ports) == 0 {
		noSimulatorsMessage()
	} else {
		fmt.Printf("Connected to the simulator on port %d\n", port)
		fmt.Printf("Use 'port', or 'next', or 'prev' to select a different simulator.\n")
	}
	fmt.Println("Enter 'help' for a list of commands.")

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	homeDir := usr.HomeDir

	// Construct the path for the history file
	historyFile := filepath.Join(homeDir, ".simtalk_history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "simtalk> ",
		HistoryFile:       historyFile,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		prompt := "simtalk"
		if app.pidx >= 0 {
			prompt += fmt.Sprintf(" simulator@%d", app.ports[app.pidx])
		}
		rl.SetPrompt(fmt.Sprintf("%s> ", prompt))

		line, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}

		text := line
		args := strings.Split(text, " ")
		trimmedText := strings.TrimSpace(text)
		switch args[0] {
		case "port":
			if app.pidx < 0 {
				noSimulatorsMessage()
				continue
			}
			if len(args) != 2 {
				noSimulatorsMessage()
				continue
			}
			port, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("Invalid port number")
				continue
			}
			fmt.Printf("Switching to simulator on port %d\n", port)
			found := false
			for i := 0; i < len(app.ports); i++ {
				if app.ports[i] == port {
					app.pidx = i
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("No simulator processes running on port %d.\n", port)
				fmt.Printf("Currently selected simulator process is listening on port %d.\n", app.ports[app.pidx])
				continue
			}
			continue

		case "ports":
			if len(app.ports) == 0 {
				noSimulatorsMessage()
				continue
			}
			fmt.Printf("Available ports:\n")
			for i := 0; i < len(app.ports); i++ {
				fmt.Printf("  %d\n", app.ports[i])
			}
			continue

		case "next":
			if app.pidx < 0 {
				noSimulatorsMessage()
				continue
			}
			app.pidx = (app.pidx + 1) % len(app.ports)
			continue

		case "prev":
			if app.pidx < 0 {
				noSimulatorsMessage()
				continue
			}
			app.pidx = (app.pidx - 1 + len(app.ports)) % len(app.ports)
			continue

		case "rescan":
			rescan()
			continue

		case "help":
			fmt.Println("Available commands:")
			for _, cmd := range commands {
				fmt.Printf("- %s : %s\n", cmd.Name, cmd.Description)
			}

		case "quit":
			fallthrough
		case "exit":
			os.Exit(0)
		}

		//-------------------------------------
		// SEND THE COMMAND TO SIMULATOR
		//-------------------------------------
		if app.pidx < 0 {
			noSimulatorsMessage()
			continue
		}
		baseURL := fmt.Sprintf("http://localhost:%d", app.ports[app.pidx])
		response, err := sendCommand(baseURL, trimmedText)
		if err != nil {
			if strings.Contains(err.Error(), "connect: connection refused") {
				fmt.Printf("A simulator is no longer running on port %d. Rescanning...\n", app.ports[app.pidx])
				rescan()
				continue
			}
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
			"               Estimated completion: %s\n"+
			"                                SID: %d\n",
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
		status.SID,
	)
}

func rescan() {
	app.pidx = -1
	app.ports = scanPorts(8090, 8100)
	if len(app.ports) == 0 {
		fmt.Printf("No simulators appear to be running on this computer\n")
		fmt.Printf("Use 'rescan' to rescan the list of available ports\n")
		return
	}
	app.pidx = 0
	fmt.Printf("Fount %d simulators running on ports these ports:\n", len(app.ports))
	for i := 0; i < len(app.ports); i++ {
		fmt.Printf("  %d\n", app.ports[i])
	}
	fmt.Printf("Now talking to simulator on port %d\n", app.ports[app.pidx])
}

func noSimulatorsMessage() {
	fmt.Printf("No simulators appear to be running on this computer\n")
	fmt.Printf("Use 'rescan' to rescan the list of available ports\n")
}
