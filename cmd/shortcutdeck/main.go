package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pnanadikar/shortcutdeck/internal/scheduler"
	"github.com/pnanadikar/shortcutdeck/internal/server"
	"github.com/pnanadikar/shortcutdeck/internal/store"
)

const (
	defaultPort   = 7432
	defaultDBName = "cards.db"
)

type runtimeConfig struct {
	port     int
	dbPath   string
	autoOpen bool
}

func main() {
	logger := newAppLogger(os.Stderr)

	cfg, err := parseRuntimeConfig(os.Args[1:], os.UserConfigDir)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		logger.Fatal(fatalLogMessage("resolve runtime config", err))
	}

	if err := os.MkdirAll(filepath.Dir(cfg.dbPath), 0o755); err != nil {
		logger.Fatal(fatalLogMessage("create db directory", err))
	}

	st, err := store.New(cfg.dbPath)
	if err != nil {
		logger.Fatal(formatStoreOpenError(cfg.dbPath, err))
	}

	srv := server.NewWithLogger(st, scheduler.SM2{}, logger)
	url := serverURL(cfg.port)
	httpServer := &http.Server{
		Addr:    listenAddr(cfg.port),
		Handler: srv.Handler(),
	}

	if cfg.autoOpen {
		go func() {
			time.Sleep(300 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				logger.Print(browserOpenWarningMessage(url, err))
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- httpServer.ListenAndServe()
	}()

	logger.Print(startupLogMessage(url, cfg.dbPath))

	select {
	case <-ctx.Done():
		_, _ = os.Stderr.WriteString("\n")
		logger.Print(shutdownLogMessage())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Fatal(fatalLogMessage("shutdown server", err))
		}
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(fatalLogMessage("listen and serve", err))
		}
	}
}

func newAppLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.LstdFlags)
}

func parseRuntimeConfig(args []string, userConfigDir func() (string, error)) (runtimeConfig, error) {
	defaultDBPath, err := resolveDefaultDBPath(userConfigDir)
	if err != nil {
		return runtimeConfig{}, err
	}

	fs, cfg := newRuntimeFlagSet(defaultDBPath, os.Stderr)

	if err := fs.Parse(normalizeRuntimeArgs(args)); err != nil {
		return runtimeConfig{}, err
	}

	if fs.NArg() > 0 {
		return runtimeConfig{}, errors.New("unexpected positional arguments; quote --db paths that contain spaces")
	}

	if cfg.port < 1 || cfg.port > 65535 {
		return runtimeConfig{}, errors.New("port must be between 1 and 65535")
	}

	if err := validateDBPath(cfg.dbPath); err != nil {
		return runtimeConfig{}, err
	}

	return *cfg, nil
}

func normalizeRuntimeArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == "--auto-open" || arg == "-auto-open") && i+1 < len(args) {
			if _, err := strconv.ParseBool(args[i+1]); err == nil {
				normalized = append(normalized, arg+"="+args[i+1])
				i++
				continue
			}
		}
		normalized = append(normalized, arg)
	}

	return normalized
}

func newRuntimeFlagSet(defaultDBPath string, output io.Writer) (*flag.FlagSet, *runtimeConfig) {
	fs := flag.NewFlagSet("shortcutdeck", flag.ContinueOnError)
	fs.SetOutput(output)

	cfg := &runtimeConfig{}
	fs.IntVar(&cfg.port, "port", defaultPort, "port for the local web server")
	fs.StringVar(&cfg.dbPath, "db", defaultDBPath, "path to the SQLite database file")
	fs.BoolVar(&cfg.autoOpen, "auto-open", true, "open the browser automatically on startup")
	fs.Usage = func() {
		_, _ = fmt.Fprintf(output, "Usage of %s:\n", fs.Name())
		_, _ = fmt.Fprintln(output, "  -auto-open")
		_, _ = fmt.Fprintln(output, "    \topen the browser automatically on startup (default true)")
		_, _ = fmt.Fprintf(output, "  -db string\n    \tpath to the SQLite database file (default %q)\n", defaultDBPath)
		_, _ = fmt.Fprintf(output, "  -port int\n    \tport for the local web server (default %d)\n", defaultPort)
	}

	return fs, cfg
}

func resolveDefaultDBPath(userConfigDir func() (string, error)) (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "shortcutdeck", defaultDBName), nil
}

func validateDBPath(dbPath string) error {
	if strings.TrimSpace(dbPath) == "" {
		return errors.New("database path must not be empty")
	}

	if strings.HasSuffix(dbPath, string(os.PathSeparator)) {
		return errors.New("database path must include a filename, not just a directory")
	}

	info, err := os.Stat(dbPath)
	if err == nil && info.IsDir() {
		return errors.New("database path must point to a file, not a directory")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check database path: %w", err)
	}

	return nil
}

func listenAddr(port int) string {
	return ":" + strconv.Itoa(port)
}

func serverURL(port int) string {
	return "http://127.0.0.1:" + strconv.Itoa(port)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func startupLogMessage(url, dbPath string) string {
	return fmt.Sprintf("shortcutdeck listening on %s (db: %s)", url, dbPath)
}

func browserOpenWarningMessage(url string, err error) string {
	return fmt.Sprintf("warning: could not open browser for %s: %v", url, err)
}

func shutdownLogMessage() string {
	return "shortcutdeck shutting down"
}

func fatalLogMessage(action string, err error) string {
	return fmt.Sprintf("%s: %v", action, err)
}

func formatStoreOpenError(dbPath string, err error) error {
	errText := strings.ToLower(err.Error())
	if strings.Contains(errText, "unable to open database file") || strings.Contains(errText, "out of memory (14)") {
		return fmt.Errorf("open store: unable to open database at %q; check that --db points to a writable file path, not a directory", dbPath)
	}

	return fmt.Errorf("open store: %w", err)
}
