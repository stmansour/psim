package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/stmansour/psim/util"
)

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
	URL                    string
	MachineID              string
	WorkingDirectory       string
}

// ShortResponse represents the response from the /stop endpoint
type ShortResponse struct {
	Status  string
	Message string
	ID      int64
}

func startHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/stopsim", handleStop)

	app.basePort = 8090
	app.maxPort = 8100
	server := &http.Server{Handler: mux}

	//---------------------------------
	// FIND A LISTENER PORT...
	//---------------------------------
	listener, err := findAvailablePort(ctx)
	if err != nil {
		return err
	}

	//------------------------------------------------------------------------
	// Start serving in a separate goroutine to allow for graceful shutdown
	//------------------------------------------------------------------------
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	//------------------------------------------------------------------------
	// Now that we have our port, get the full id and contact info for this machine
	//------------------------------------------------------------------------
	myID, err := util.GetMachineUUID()
	if err != nil {
		fmt.Printf("Error getting machine UUID: %v\n", err)
		myID = "uknown"
	}
	app.MachineID = myID
	addrlist, err := util.GetNetworkInfo()
	if err != nil || app.SID == 0 {
		fmt.Printf("Simtalk port: %d\n", app.Simtalkport)
	} else {
		for _, addr := range addrlist {
			if addr.IPAddress == "127.0.0.1" {
				continue
			}
			//-----------------------------------------------------------------------------
			// this is the URL that we can depend on, the host name may not be resolvable
			//-----------------------------------------------------------------------------
			app.URL = fmt.Sprintf("http://%s:%d", addr.IPAddress, app.Simtalkport)
			fmt.Printf("Simtalk address: %s\n", app.URL)
		}
	}

	//----------------------------------------------------------------------------------------
	// Wait for the context to be canceled (simulation done), then shut down the server
	//----------------------------------------------------------------------------------------
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func findAvailablePort(ctx context.Context) (net.Listener, error) {
	for port := app.basePort; port <= app.maxPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			app.Simtalkport = port
			return listener, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("could not find an available port between %d and %d", app.basePort, app.maxPort)
}

func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())

	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds%60)
	} else if hours < 24 {
		return fmt.Sprintf("%d hours, %d minutes, %d seconds", hours, minutes%60, seconds%60)
	} else {
		days := hours / 24
		return fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours%24, minutes%60, seconds%60)
	}
}

func strElapsedTime(start, end time.Time) string {
	return formatDuration(end.Sub(start))
}

// func estimateFinish() (int, time.Duration, time.Time) {
// 	totalGens := app.cfg.LoopCount * app.cfg.Generations
// 	completedGens := (app.sim.LoopsCompleted * app.cfg.Generations) + app.sim.GensCompleted
// 	gensRemaining := totalGens - completedGens

// 	timePerGen := app.sim.TrackingGenStop.Sub(app.sim.TrackingGenStart) // Calculate the time taken for the last generation
// 	estimatedTimeRemaining := timePerGen * time.Duration(gensRemaining) // Calculate the estimated time remaining
// 	estimatedCompletionTime := time.Now().Add(estimatedTimeRemaining)   // Calculate the estimated completion time

// 	return completedGens, estimatedTimeRemaining, estimatedCompletionTime

// }

// endTimeEstimator calculates the estimated completion time
// based on the provided seconds per generation, number of loops, and
// number of generations per loop.
// RETURNS:
//
//	estimated time.Time of completion
//	estimated time.Duration of remaining time
//
// --------------------------------------------------------------------------
func endTimeEstimator() (int, time.Duration, time.Duration, time.Time) {
	timePerGen := app.sim.TrackingGenStop.Sub(app.sim.TrackingGenStart)   //the time taken for the last generation
	totalGenerations := app.cfg.LoopCount * app.cfg.Generations           // TODO: we need a different formula if GenDur is set
	remainingGenerations := totalGenerations - app.sim.GensCompleted      // how many generations are left to simulate
	remainingDuration := timePerGen * time.Duration(remainingGenerations) // estimated time remaining
	currentTime := time.Now()                                             // what time is it now?
	estimatedCompletionTime := currentTime.Add(remainingDuration)         // Add the duration to now

	return totalGenerations, timePerGen, remainingDuration, estimatedCompletionTime
}

// handleStatus returns the status of the simulation. Times are in UTC
func handleStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("**** HTTP STATUS HANDLER has been entered\n")
	timeElapsed := strElapsedTime(app.ProgramStarted, time.Now())

	_, timePerGen, estimatedTimeRemaining, estimatedCompletionTime := endTimeEstimator()

	status := SimulatorStatus{
		ProgramStarted:         app.ProgramStarted.In(time.UTC).Format(time.RFC3339),
		RunDuration:            timeElapsed,
		ConfigFile:             app.cfg.ConfigFilename,
		SimulationDateRange:    time.Time(app.cfg.DtStart).Format("Jan 2, 2006") + " - " + time.Time(app.cfg.DtStop).Format("Jan 2, 2006"),
		PopulationSize:         app.cfg.PopulationSize,
		LoopCount:              app.cfg.LoopCount,
		GenerationsRequested:   app.cfg.Generations,
		CompletedLoops:         app.sim.LoopsCompleted,
		CompletedGenerations:   app.sim.GensCompleted,
		ElapsedTimeLastGen:     util.ElapsedDuration(timePerGen),
		EstimatedTimeRemaining: formatDuration(estimatedTimeRemaining),
		EstimatedCompletion:    estimatedCompletionTime.In(time.UTC).Format(time.RFC3339),
		SID:                    app.SID,
		URL:                    app.URL,
		MachineID:              app.MachineID,
		WorkingDirectory:       app.WorkingDirectory,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode status", http.StatusInternalServerError)
	}
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("**** HTTP STATUS HANDLER has been entered\n")
	// Implementation
	app.sim.Cfg.LoopCount = 1
	app.sim.Cfg.Generations = 1

	response := ShortResponse{
		Status:  "Success",
		Message: "Stopping after current generation",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
