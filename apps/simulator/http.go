package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
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

func handleStatus(w http.ResponseWriter, r *http.Request) {
	dtStart := time.Time(app.cfg.DtStart)
	dtStop := time.Time(app.cfg.DtStop)
	fmt.Fprintf(w, "SIMULATOR STATUS\n")
	fmt.Fprintf(w, "                 Start, Stop: %s - %s\n", dtStart.Format("Jan 2, 2006"), dtStop.Format("Jan 2, 2006"))
	fmt.Fprintf(w, "                 Generations: %d\n", app.cfg.Generations)
	fmt.Fprintf(w, "                   LoopCount: %d\n", app.cfg.LoopCount)
	fmt.Fprintf(w, "           total Generations: %d\n", app.cfg.Generations*app.cfg.LoopCount)
	fmt.Fprintf(w, "      Generation in progress: %d\n", app.sim.GensCompleted)
	fmt.Fprintf(w, "Elapsed time last generation:%s\n", app.sim.GenerationSimTime)
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	// ImpleCurrently running generation: %d of %d\n",app.sim.GensCompleted,app.cfg.Generationson
	app.sim.Cfg.LoopCount = 1
	app.sim.Cfg.Generations = 1

	fmt.Fprintf(w, "Stopping after current generation...\n")
}
