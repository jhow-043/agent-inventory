// Package main is the entry point for the Windows Inventory Agent.
// It supports running as a Windows service or in foreground mode.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"inventario/agent/internal/client"
	"inventario/agent/internal/collector"
	"inventario/agent/internal/config"
	"inventario/agent/internal/token"
)

const (
	serviceName    = "InventoryAgent"
	serviceDisplay = "Inventory Agent"
	serviceDesc    = "Windows IT Asset Inventory Agent"
	version        = "1.0.0"
)

func main() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "version":
			fmt.Printf("inventory-agent v%s\n", version)
			return
		case "install":
			installService()
			return
		case "uninstall":
			uninstallService()
			return
		case "start":
			startService()
			return
		case "stop":
			stopService()
			return
		case "run":
			runForeground()
			return
		case "collect":
			runCollectOnly()
			return
		}
	}

	// If launched without arguments, check whether we are running as a Windows service.
	isService, err := svc.IsWindowsService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine service status: %v\n", err)
		os.Exit(1)
	}

	if isService {
		runWindowsService()
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Printf("Inventory Agent v%s\n\n", version)
	fmt.Println("Commands:")
	fmt.Println("  collect     Collect inventory and print JSON (no server needed)")
	fmt.Println("  install     Install as Windows service")
	fmt.Println("  uninstall   Remove Windows service")
	fmt.Println("  start       Start the Windows service")
	fmt.Println("  stop        Stop the Windows service")
	fmt.Println("  run         Run in foreground (debug mode)")
	fmt.Println("  version     Show version")
	fmt.Println()
	fmt.Println("Flags for 'run':")
	fmt.Println("  -config string  Path to config.json (default: next to executable)")
}

// ---------------------------------------------------------------------------
// Collect-only mode (dry run — no API needed)
// ---------------------------------------------------------------------------

func runCollectOnly() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	coll := collector.New(logger)

	inventory, err := coll.Collect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "collection failed: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

// ---------------------------------------------------------------------------
// Windows Service
// ---------------------------------------------------------------------------

// agentService implements the svc.Handler interface for the Windows SCM.
type agentService struct{}

func (s *agentService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for service control requests in a separate goroutine.
	go func() {
		for {
			select {
			case c := <-r:
				switch c.Cmd {
				case svc.Interrogate:
					changes <- c.CurrentStatus
				case svc.Stop, svc.Shutdown:
					changes <- svc.Status{State: svc.StopPending}
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	runAgent(ctx, "")

	return false, 0
}

func runWindowsService() {
	if err := svc.Run(serviceName, &agentService{}); err != nil {
		fmt.Fprintf(os.Stderr, "service run failed: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// Foreground (debug) mode
// ---------------------------------------------------------------------------

func runForeground() {
	var configPath string
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.StringVar(&configPath, "config", "", "Path to config.json")
	_ = fs.Parse(os.Args[2:])

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	runAgent(ctx, configPath)
}

// ---------------------------------------------------------------------------
// Core agent loop
// ---------------------------------------------------------------------------

func runAgent(ctx context.Context, configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	logger := setupLogger(cfg.LogLevel)
	logger.Info("starting inventory agent", "version", version)

	store := token.NewStore(cfg.DataDir)
	coll := collector.New(logger)
	apiClient := client.New(cfg.ServerURL, cfg.InsecureSkipVerify, logger)

	// Load existing token if available.
	tok, err := store.Load()
	if err != nil {
		logger.Error("failed to load saved token", "error", err)
	}

	// Run initial inventory cycle immediately.
	runCycle(ctx, cfg, logger, store, coll, apiClient, &tok)

	// Schedule periodic cycles.
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("agent shutting down")
			return
		case <-ticker.C:
			runCycle(ctx, cfg, logger, store, coll, apiClient, &tok)
		}
	}
}

func runCycle(
	ctx context.Context,
	cfg *config.Config,
	logger *slog.Logger,
	store *token.Store,
	coll *collector.Collector,
	apiClient *client.Client,
	tok *string,
) {
	logger.Info("starting inventory cycle")

	inventory, err := coll.Collect()
	if err != nil {
		logger.Error("inventory collection failed", "error", err)
		return
	}

	// Enroll if we have no token yet.
	if *tok == "" {
		logger.Info("no device token found, enrolling")
		resp, err := apiClient.Enroll(ctx, cfg.EnrollmentKey, inventory.Hostname, inventory.SerialNumber)
		if err != nil {
			logger.Error("enrollment failed", "error", err)
			return
		}
		*tok = resp.Token
		if err := store.Save(*tok); err != nil {
			logger.Error("failed to persist token", "error", err)
		}
		logger.Info("enrolled successfully", "device_id", resp.DeviceID)
	}

	// Submit the inventory snapshot.
	apiClient.SetToken(*tok)
	if err := apiClient.SubmitWithRetry(ctx, inventory, 5); err != nil {
		logger.Error("inventory submission failed", "error", err)

		// If we get a 401/403, the token is invalid — clear it so we re-enroll next cycle.
		if client.IsAuthError(err) {
			logger.Info("token appears invalid, clearing for re-enrollment")
			*tok = ""
			_ = store.Delete()
		}
		return
	}

	logger.Info("inventory submitted successfully")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
}

// ---------------------------------------------------------------------------
// Service install / uninstall / start / stop
// ---------------------------------------------------------------------------

func installService() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to service manager: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		fmt.Println("service already exists")
		return
	}

	s, err = m.CreateService(serviceName, exe, mgr.Config{
		DisplayName: serviceDisplay,
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create service: %v\n", err)
		os.Exit(1)
	}
	s.Close()
	fmt.Println("service installed successfully")
}

func uninstallService() {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to service manager: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not found: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete service: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("service uninstalled successfully")
}

func startService() {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to service manager: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not found: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start service: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("service started")
}

func stopService() {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to service manager: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not found: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if _, err := s.Control(svc.Stop); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop service: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("service stop signal sent")
}
