package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/m5rcel-vibecodes/syspeek/internal/collector"
	"github.com/m5rcel-vibecodes/syspeek/internal/config"
	"github.com/m5rcel-vibecodes/syspeek/internal/ui"
)

const version = "1.0.0"

type cliOptions struct {
	Command       string
	JSON          bool
	Watch         bool
	WatchInterval float64
	ConfigPath    string
	NoColor       bool
	Compact       bool
	ShowHelp      bool
	ShowVersion   bool
}

func parseArgs(args []string) *cliOptions {
	opts := &cliOptions{
		WatchInterval: 2.0,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h" || arg == "help":
			opts.ShowHelp = true
			return opts
		case arg == "--version" || arg == "-v" || arg == "version":
			opts.ShowVersion = true
			return opts
		case arg == "--json":
			opts.JSON = true
		case arg == "--no-color":
			opts.NoColor = true
		case arg == "--compact":
			opts.Compact = true
		case arg == "-c" || arg == "--config":
			if i+1 < len(args) {
				opts.ConfigPath = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--config="):
			opts.ConfigPath = strings.TrimPrefix(arg, "--config=")
		case arg == "-w" || arg == "--watch":
			opts.Watch = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				if sec, err := strconv.ParseFloat(args[i+1], 64); err == nil && sec > 0 {
					opts.WatchInterval = sec
					i++
				}
			}
		case strings.HasPrefix(arg, "--watch="):
			opts.Watch = true
			secStr := strings.TrimPrefix(arg, "--watch=")
			if sec, err := strconv.ParseFloat(secStr, 64); err == nil && sec > 0 {
				opts.WatchInterval = sec
			}
		case strings.HasPrefix(arg, "-w="):
			opts.Watch = true
			secStr := strings.TrimPrefix(arg, "-w=")
			if sec, err := strconv.ParseFloat(secStr, 64); err == nil && sec > 0 {
				opts.WatchInterval = sec
			}
		case !strings.HasPrefix(arg, "-") && opts.Command == "":
			opts.Command = strings.ToLower(arg)
		}
		i++
	}

	return opts
}

func printHelp() {
	fmt.Println(`syspeek - Fast, dependency-light system diagnostic tool

USAGE:
    syspeek [COMMAND] [FLAGS]

COMMANDS:
    (default)       Display the unified system dashboard
    cpu             Detailed CPU architecture, cores, usage & thermals
    memory          Detailed RAM and Swap memory statistics
    disk, storage   Mounted filesystems, capacity and usage breakdown
    gpu             Graphics accelerators, VRAM and driver status
    network         Network interfaces, IP addresses and transfer stats
    services        System service manager (systemd, OpenRC, launchd) status
    processes, top  Process overview sorted by CPU and memory usage
    version         Print version information
    help            Print this help message

FLAGS:
    --json          Output pure machine-readable JSON
    -w, --watch [N] Periodically refresh output (default: 2 seconds)
    -c, --config    Path to custom configuration file
    --compact       Render in compact mode with reduced vertical padding
    --no-color      Disable ANSI color output (also respects NO_COLOR=1)
    -v, --version   Print version information
    -h, --help      Print this help message

EXAMPLES:
    syspeek                     # Show system dashboard
    syspeek --watch             # Refresh dashboard every 2 seconds
    syspeek --watch 1           # Refresh dashboard every second
    syspeek --json              # Output complete snapshot as JSON
    syspeek cpu --json          # Output CPU diagnostics as JSON
    syspeek top                 # Live refreshing process monitor
    syspeek network             # Show all network interfaces`)
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func executeCommand(cmd string, cfg *config.Config, isJSON bool) {
	switch cmd {
	case "", "dashboard":
		snap, err := collector.CollectSnapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting snapshot: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(snap)
		} else {
			ui.RenderDashboard(os.Stdout, snap, cfg)
		}

	case "cpu":
		cpu, err := collector.CollectCPU()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting CPU metrics: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(cpu)
		} else {
			ui.RenderCPUView(os.Stdout, cpu, cfg)
		}

	case "memory", "mem":
		mem, err := collector.CollectMemory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting memory metrics: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(mem)
		} else {
			ui.RenderMemoryView(os.Stdout, mem, cfg)
		}

	case "disk", "storage":
		disks, err := collector.CollectDisks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting disk metrics: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(disks)
		} else {
			ui.RenderDiskView(os.Stdout, disks, cfg)
		}

	case "gpu":
		gpus, err := collector.CollectGPUs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting GPU metrics: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(gpus)
		} else {
			ui.RenderGPUView(os.Stdout, gpus, cfg)
		}

	case "network", "net":
		ifaces, err := collector.CollectNetwork()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting network metrics: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(ifaces)
		} else {
			ui.RenderNetworkView(os.Stdout, ifaces, cfg)
		}

	case "services", "service":
		services, err := collector.CollectServices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting services: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(services)
		} else {
			ui.RenderServicesView(os.Stdout, services, cfg)
		}

	case "processes", "top", "ps":
		limit := 25
		if cfg.Compact {
			limit = 15
		}
		procs, err := collector.CollectProcesses(limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting processes: %v\n", err)
			os.Exit(1)
		}
		if isJSON {
			printJSON(procs)
		} else {
			ui.RenderProcessesView(os.Stdout, procs, cfg)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nRun 'syspeek --help' for usage.\n", cmd)
		os.Exit(1)
	}
}

func main() {
	opts := parseArgs(os.Args[1:])

	if opts.ShowHelp {
		printHelp()
		return
	}

	if opts.ShowVersion {
		fmt.Printf("syspeek v%s\n", version)
		return
	}

	// Load configuration
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
		cfg = config.DefaultConfig()
	}

	// CLI flags override config
	if opts.NoColor {
		cfg.Colors = false
	}
	if opts.Compact {
		cfg.Compact = true
	}

	// Top command defaults to interactive watch mode unless JSON was requested
	if opts.Command == "top" && !opts.JSON {
		opts.Watch = true
	}

	if !opts.Watch {
		executeCommand(opts.Command, cfg, opts.JSON)
		return
	}

	// Watch / Live Refresh Mode
	interval := time.Duration(opts.WatchInterval * float64(time.Second))
	if interval < 200*time.Millisecond {
		interval = 200 * time.Millisecond
	}

	// Signal handling for clean exit and cursor restoration
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if !opts.JSON {
		ui.HideCursor()
		defer func() {
			ui.ShowCursor()
			fmt.Println()
		}()
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial render
	if !opts.JSON {
		ui.ClearScreen()
	}
	executeCommand(opts.Command, cfg, opts.JSON)

	for {
		select {
		case <-sigChan:
			return
		case <-ticker.C:
			if !opts.JSON {
				ui.ClearScreen()
			}
			executeCommand(opts.Command, cfg, opts.JSON)
		}
	}
}
