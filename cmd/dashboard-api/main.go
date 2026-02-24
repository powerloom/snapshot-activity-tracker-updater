package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"p2p-debugger/internal/dashboard"
	trackerRedis "p2p-debugger/redis"
)

//go:embed frontend/dist
var frontendFS embed.FS

var (
	dataMarket = flag.String("data-market", os.Getenv("DATA_MARKET_ADDRESS"), "Data market contract address")
	dumpDir    = flag.String("dump-dir", os.Getenv("TALLY_DUMP_DIR"), "Directory containing tally dump files")
	port       = flag.String("port", os.Getenv("DASHBOARD_PORT"), "HTTP port to listen on")
	devMode    = flag.Bool("dev", os.Getenv("DASHBOARD_DEV_MODE") == "true", "Enable development mode (serve from filesystem)")
)

func main() {
	flag.Parse()

	// Set default port
	if *port == "" {
		*port = "8080"
	}

	// Set default dump directory
	if *dumpDir == "" {
		*dumpDir = "./tallies"
	}

	// Validate data market address
	if *dataMarket == "" {
		log.Fatal("DATA_MARKET_ADDRESS environment variable or --data-market flag is required")
	}

	// Initialize Redis client
	redisClient, err := trackerRedis.NewRedisClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	trackerRedis.RedisClient = redisClient
	log.Println("Connected to Redis")

	// Create dashboard server
	server := dashboard.NewServer(*dataMarket, *dumpDir)

	// Setup static file serving
	if *devMode {
		// Development mode: serve from filesystem
		log.Println("Running in development mode - serving frontend from filesystem")
		spaHandler := dashboard.NewSpaHandler("../../frontend/dist")
		server.SetIndexHandler(spaHandler)
	} else {
		// Production mode: serve from embedded files
		log.Println("Running in production mode - serving embedded frontend")
		dashboard.WithStaticFiles(frontendFS)(server)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%s", *port)
		serverErr <- server.ListenAndServe(addr)
	}()

	log.Printf("Dashboard API server listening on :%s", *port)
	log.Printf("Dashboard UI available at http://localhost:%s", *port)
	log.Printf("API health check at http://localhost:%s/api/health", *port)
	log.Printf("Network topology at http://localhost:%s/api/network/topology", *port)

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		log.Println("Shutting down dashboard server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdownCtx
	case err := <-serverErr:
		log.Fatalf("Dashboard server error: %v", err)
	}

	log.Println("Dashboard server stopped")
}
