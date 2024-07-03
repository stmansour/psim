package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stmansour/psim/util"
)

// Command represents the structure of a command
type Command struct {
	Command  string
	Username string
	Data     json.RawMessage
}

// SendStatusUpdate periodically sends progress information on this simulation to the dispatcher.
//
//		MachineID
//		URL
//		CPUS
//		Memory
//		DtEstimate
//		DtCompleted
//
//	 Supply 'completed' only when the simulation is complete.
//
// -------------------------------------------------------------------------------------------
func SendStatusUpdate(completed *time.Time) error {
	cmd := Command{
		Command:  "UpdateItem",
		Username: "simulator",
	}

	cmdDataStruct := struct {
		SID             int64
		MachineID       string
		CPUs            int
		Memory          string
		CPUArchitecture string
		Availability    string
		DtEstimate      string
		DtCompleted     string
	}{}

	var err error
	cmdDataStruct.SID = app.SID
	cmdDataStruct.MachineID = app.MachineID
	cmdDataStruct.CPUs = 10                 // TODO: get real value
	cmdDataStruct.Memory = "64GB"           // TODO: get real value
	cmdDataStruct.CPUArchitecture = "ARM64" // TODO: get real value
	if completed != nil {
		cmdDataStruct.DtCompleted = completed.Format(time.RFC822Z)
	} else {
		completedGens, _, estimatedCompletionTime := estimateFinish()
		if completedGens == 0 {
			return nil // nothing to report.  We need at least 1 generation to be completed
		}
		cmdDataStruct.DtEstimate = estimatedCompletionTime.Format(time.RFC822Z)
	}

	dataBytes, err := json.Marshal(cmdDataStruct)
	if err != nil {
		return fmt.Errorf("failed to marshal book request: %v", err)
	}
	cmd.Data = json.RawMessage(dataBytes)
	sendCmdData, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal book request: %v", err)
	}

	if app.HTTPHdrsDbg {
		util.PrintHexAndASCII(sendCmdData, len(sendCmdData))
	}

	// ----------------------------------------
	// Create the URL to the dispatcher
	// ----------------------------------------
	url := fmt.Sprintf("%scommand", app.DispatcherURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(sendCmdData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// ----------------------------------------
	// DEBUG:Print request headers
	// ----------------------------------------
	if app.HTTPHdrsDbg {
		fmt.Println("Request Headers:")
		for k, v := range req.Header {
			fmt.Printf("%s: %s\n", k, v)
		}
	}

	// ----------------------------------------
	// Send the request
	// ----------------------------------------
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send book request: %v", err)
	}
	defer resp.Body.Close()

	// ----------------------------------------
	// Read the response
	// ----------------------------------------
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	// ----------------------------------------
	// DEBUG:
	// ----------------------------------------
	if app.HTTPHdrsDbg {
		util.PrintHexAndASCII(bodyBytes, len(bodyBytes))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v, body: %s", resp.StatusCode, string(bodyBytes))
	}
	var response ShortResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.Printf("(SendStatusUpdate) failed to unmarshal response: %v", err)
	}
	if strings.ToLower(response.Status) != "success" {
		log.Printf("(SendStatusUpdate) unexpected dispatcher response: status: %s, message: %s", response.Status, response.Message)
	}
	return nil
}
