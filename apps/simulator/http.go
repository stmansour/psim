package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/stmansour/psim/util"
)

func startHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/stop", handleStop)

	app.basePort = 8080
	app.maxPort = 8100
	server := &http.Server{Handler: mux}

	// Attempt to listen on a range of ports
	listener, err := findAvailablePort(ctx)
	if err != nil {
		return err
	}

	// Start serving in a separate goroutine to allow for graceful shutdown
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	log.Printf("Listening for commands on http://localhost:%d\n", app.Simtalkport)

	// Wait for the context to be canceled (simulation done), then shut down the server
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

func handleStatus(w http.ResponseWriter, r *http.Request) {
	timeElapsed := strElapsedTime(app.ProgramStarted, time.Now())

	totalGens := app.cfg.LoopCount * app.cfg.Generations
	completedGens := (app.sim.LoopsCompleted + 1) * app.sim.GensCompleted
	gensRemaining := totalGens - completedGens

	timePerGen := app.sim.TrackingGenStop.Sub(app.sim.TrackingGenStart) // Calculate the time taken for the last generation
	estimatedTimeRemaining := timePerGen * time.Duration(gensRemaining) // Calculate the estimated time remaining
	estimatedCompletionTime := time.Now().Add(estimatedTimeRemaining)   // Calculate the estimated completion time

	dtStart := time.Time(app.cfg.DtStart)
	dtStop := time.Time(app.cfg.DtStop)
	fmt.Fprintf(w, "SIMULATOR STATUS\n")
	fmt.Fprintf(w, "                    Program started: %s\n", app.ProgramStarted.Format(time.RFC1123))
	fmt.Fprintf(w, "                Run duration so far: %s\n", timeElapsed)
	fmt.Fprintf(w, "                        Config file: %s\n", app.cfg.ConfigFilename)
	fmt.Fprintf(w, "              Simulation Date Range: %s - %s\n", dtStart.Format("Jan 2, 2006"), dtStop.Format("Jan 2, 2006"))
	fmt.Fprintf(w, "LoopCount and Generations requested: loops: %d, generations: %d\n", app.cfg.LoopCount, app.cfg.Generations)
	fmt.Fprintf(w, "                          completed: loops: %d, generations: %d\n", app.sim.LoopsCompleted, app.sim.GensCompleted)
	fmt.Fprintf(w, "       Elapsed time last generation: %s\n", util.ElapsedTime(app.sim.TrackingGenStart, app.sim.TrackingGenStop))
	fmt.Fprintf(w, "           Estimated time remaining: %s\n", formatDuration(estimatedTimeRemaining))
	fmt.Fprintf(w, "               Estimated completion: %s\n", estimatedCompletionTime.Format(time.RFC1123))
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	// ImpleCurrently running generation: %d of %d\n",app.sim.GensCompleted,app.cfg.Generationson
	app.sim.Cfg.LoopCount = 1
	app.sim.Cfg.Generations = 1

	fmt.Fprintf(w, "Stopping after current generation...\n")
}
