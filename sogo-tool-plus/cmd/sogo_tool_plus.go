package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sonroyaalmerol/sogo-tool-plus/internal/lock"
	"github.com/sonroyaalmerol/sogo-tool-plus/internal/sogo"
	"github.com/sonroyaalmerol/sogo-tool-plus/internal/web"
)

func runLocked(configFile, key string, ttl time.Duration, argv []string) int {
	if len(argv) == 0 {
		log.Fatal("Missing command to run after '--' for mode 'lock'")
	}
	if key == "" {
		log.Fatal("Missing required flag: -key for mode 'lock'")
	}

	held, acquired := lock.Acquire(configFile, key, ttl)
	if !acquired {
		return 0
	}
	defer held.Release()

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		if sig, ok := <-signals; ok && cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
	}()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		log.Printf("Failed to run %q: %v", argv[0], err)
		return 1
	}

	return 0
}

func runWebServer(service *sogo.SogoService, addr string) {
	handler := web.NewWebHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/calendars/subscribe/user/", handler.HandleCalSubscribeUser)
	mux.HandleFunc("/calendars/subscribe/all", handler.HandleCalSubscribeAll)

	log.Printf("Starting web server on %s", addr)
	err := http.ListenAndServe(addr, mux)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Web server failed: %v", err)
	}
	log.Println("Web server stopped.")
}

func runCLI(service *sogo.SogoService, action, uid string, duration int) {
	log.Println("Running in CLI mode")
	var err error

	switch strings.ToLower(action) {
	case "cal-subscribe-user":
		if uid == "" {
			log.Fatal(
				"Missing required flag: -uid for action 'cal-subscribe-user'",
			)
		}
		log.Printf("Action: Subscribe User '%s'", uid)
		err = service.CalSubscribeUser(uid)
	case "cal-subscribe-all":
		log.Println("Action: Subscribe All Users")
		err = service.CalSubscribeAll()
	case "expire-sessions-creation":
		if duration == 0 {
			log.Fatal(
				"Missing required flag: -duration for action 'expire-sessions-creation'. Must be non-zero.",
			)
		}
		log.Println("Action: Expire sessions by creation date")
		err = service.DeleteSessionsByCreation(time.Duration(duration) * time.Minute)
	case "":
		log.Fatal(
			"Missing required flag: -action (e.g., 'cal-subscribe-user', 'cal-subscribe-all')",
		)
	default:
		log.Fatalf(
			"Invalid action: %s. Use 'cal-subscribe-user', 'cal-subscribe-all', or 'expire-sessions-creation'.",
			action,
		)
	}

	if err != nil {
		log.Fatalf("CLI action failed: %v", err)
	}

	log.Println("CLI action completed successfully.")
}

func main() {
	mode := flag.String(
		"mode",
		"cli",
		"Operation mode: 'cli', 'server', or 'lock'",
	)
	action := flag.String(
		"action",
		"",
		"CLI action: 'cal-subscribe-user', 'cal-subscribe-all', or 'expire-sessions-creation'",
	)
	uid := flag.String("uid", "", "User ID for 'cal-subscribe-user' action")
	sessionDuration := flag.Int("duration", 0, "Max duration for sessions since creation date in minutes (must be greater than zero)")
	configFile := flag.String(
		"config",
		"/etc/sogo/sogo.conf",
		"Path to sogo.conf",
	)
	serverAddr := flag.String(
		"addr",
		":8008",
		"Address for the web server to listen on (e.g., :8008)",
	)
	lockKey := flag.String(
		"key",
		"",
		"Lock name for mode 'lock'",
	)
	lockTTL := flag.Int(
		"ttl",
		300,
		"Lock expiry in seconds for mode 'lock'",
	)

	flag.Parse()

	if strings.EqualFold(*mode, "lock") {
		os.Exit(runLocked(
			*configFile,
			*lockKey,
			time.Duration(*lockTTL)*time.Second,
			flag.Args(),
		))
	}

	service, err := sogo.NewSogoService(*configFile)
	if err != nil {
		log.Fatalf("Failed to initialize Sogo service: %v", err)
	}
	defer service.Close()

	switch strings.ToLower(*mode) {
	case "cli":
		runCLI(service, *action, *uid, *sessionDuration)
	case "server":
		runWebServer(service, *serverAddr)
	default:
		log.Fatalf("Invalid mode: %s. Use 'cli', 'server', or 'lock'.", *mode)
	}
}
