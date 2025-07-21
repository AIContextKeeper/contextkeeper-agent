package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/carsor007/contextkeeper-agent/internal/agent"
	"github.com/carsor007/contextkeeper-agent/pkg/types"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	var (
		configFile    = flag.String("config", "", "Path to config file")
		showUsage     = flag.Bool("usage", false, "Show current usage information")
		showVer       = flag.Bool("version", false, "Show version information")
		showDashboard = flag.Bool("dashboard", false, "Open ContextKeeper dashboard")
		showSession   = flag.Bool("session", false, "Show session information")
		daemon        = flag.Bool("daemon", false, "Run as daemon")
		logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Show version
	if *showVer {
		fmt.Printf("ContextKeeper Agent\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit:  %s\n", commit)
		fmt.Printf("Date:    %s\n", date)
		return
	}

	// Setup logging
	setupLogging(*logLevel)

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create agent
	a, err := agent.New(config)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Show usage information
	if *showUsage {
		usage, err := a.GetUsageInfo()
		if err != nil {
			log.Fatalf("Failed to get usage info: %v", err)
		}
		
		fmt.Printf("ContextKeeper Usage Information\n")
		fmt.Printf("==============================\n")
		fmt.Printf("Current sessions: %d\n", usage.Current)
		if usage.Limit > 0 {
			fmt.Printf("Limit: %d\n", usage.Limit)
			fmt.Printf("Usage: %d%%\n", usage.Percentage)
			fmt.Printf("Remaining: %d\n", usage.Limit-usage.Current)
		} else {
			fmt.Printf("Limit: Unlimited (Pro)\n")
		}
		return
	}

	// Show session information
	if *showSession {
		a.ShowSessionInfo()
		return
	}

	// Open dashboard
	if *showDashboard {
		if err := a.ShowDashboard(); err != nil {
			log.Fatalf("Failed to open dashboard: %v", err)
		}
		return
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start agent
	log.Printf("Starting ContextKeeper Agent v%s", version)
	if err := a.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// Run as daemon or interactive
	if *daemon {
		log.Printf("Running as daemon. Press Ctrl+C to stop.")
		// Wait for signal
		<-sigChan
	} else {
		log.Printf("Agent started. Press Ctrl+C to stop.")
		// Wait for signal
		<-sigChan
	}

	log.Printf("Shutting down...")
	cancel()

	// Stop agent
	if err := a.Stop(); err != nil {
		log.Printf("Error stopping agent: %v", err)
	}

	log.Printf("Agent stopped")
}

// loadConfig loads configuration from file or creates default
func loadConfig(configFile string) (*types.Config, error) {
	config := &types.Config{
		ServerURL:    "https://contextkeeper.dev",
		LocalPort:    8080,
		LogLevel:     "info",
		EnableTLS:    true,
		MaxSessions:  100,
		UploadBatch:  5,
	}

	// If config file specified, try to load it
	if configFile != "" {
		// TODO: Implement config file loading
		log.Printf("Config file loading not yet implemented, using defaults")
	}

	return config, nil
}

// setupLogging configures logging based on level
func setupLogging(level string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	switch level {
	case "debug":
		log.SetOutput(os.Stdout)
	case "info":
		log.SetOutput(os.Stdout)
	case "warn", "error":
		log.SetOutput(os.Stderr)
	default:
		log.SetOutput(os.Stdout)
	}
}