package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRuntimeConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := parseRuntimeConfig(nil, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err != nil {
		t.Fatalf("parseRuntimeConfig() error = %v", err)
	}

	if cfg.port != defaultPort {
		t.Fatalf("port = %d, want %d", cfg.port, defaultPort)
	}

	wantDBPath := filepath.Join("/tmp/config-home", "shortcutdeck", defaultDBName)
	if cfg.dbPath != wantDBPath {
		t.Fatalf("dbPath = %q, want %q", cfg.dbPath, wantDBPath)
	}

	if !cfg.autoOpen {
		t.Fatal("autoOpen = false, want true")
	}
}

func TestParseRuntimeConfigExplicitValues(t *testing.T) {
	t.Parallel()

	cfg, err := parseRuntimeConfig([]string{"--port", "8123", "--db", "/tmp/custom.db"}, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err != nil {
		t.Fatalf("parseRuntimeConfig() error = %v", err)
	}

	if cfg.port != 8123 {
		t.Fatalf("port = %d, want 8123", cfg.port)
	}

	if cfg.dbPath != "/tmp/custom.db" {
		t.Fatalf("dbPath = %q, want %q", cfg.dbPath, "/tmp/custom.db")
	}

	if !cfg.autoOpen {
		t.Fatal("autoOpen = false, want true")
	}
}

func TestParseRuntimeConfigAllowsDisablingAutoOpen(t *testing.T) {
	t.Parallel()

	cfg, err := parseRuntimeConfig([]string{"--auto-open=false"}, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err != nil {
		t.Fatalf("parseRuntimeConfig() error = %v", err)
	}

	if cfg.autoOpen {
		t.Fatal("autoOpen = true, want false")
	}
}

func TestParseRuntimeConfigAllowsDisablingAutoOpenWithSeparateValue(t *testing.T) {
	t.Parallel()

	cfg, err := parseRuntimeConfig([]string{"--auto-open", "false"}, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err != nil {
		t.Fatalf("parseRuntimeConfig() error = %v", err)
	}

	if cfg.autoOpen {
		t.Fatal("autoOpen = true, want false")
	}
}

func TestParseRuntimeConfigRejectsUnexpectedArgs(t *testing.T) {
	t.Parallel()

	_, err := parseRuntimeConfig([]string{"--db", "/tmp/Application", "Support/cards.db"}, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err == nil {
		t.Fatal("parseRuntimeConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "unexpected positional arguments") {
		t.Fatalf("error = %q, want positional argument guidance", err)
	}
}

func TestParseRuntimeConfigRejectsInvalidPort(t *testing.T) {
	t.Parallel()

	_, err := parseRuntimeConfig([]string{"--port", "70000"}, func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err == nil {
		t.Fatal("parseRuntimeConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "port must be between 1 and 65535") {
		t.Fatalf("error = %q, want invalid port message", err)
	}
}

func TestValidateDBPathRejectsDirectory(t *testing.T) {
	t.Parallel()

	dbDir := t.TempDir()
	err := validateDBPath(dbDir)
	if err == nil {
		t.Fatal("validateDBPath() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "must point to a file") {
		t.Fatalf("error = %q, want directory guidance", err)
	}
}

func TestResolveDefaultDBPath(t *testing.T) {
	t.Parallel()

	got, err := resolveDefaultDBPath(func() (string, error) {
		return "/tmp/config-home", nil
	})
	if err != nil {
		t.Fatalf("resolveDefaultDBPath() error = %v", err)
	}

	want := filepath.Join("/tmp/config-home", "shortcutdeck", defaultDBName)
	if got != want {
		t.Fatalf("resolveDefaultDBPath() = %q, want %q", got, want)
	}
}

func TestResolveDefaultDBPathPropagatesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("no config dir")
	_, err := resolveDefaultDBPath(func() (string, error) {
		return "", wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("resolveDefaultDBPath() error = %v, want %v", err, wantErr)
	}
}

func TestNewRuntimeFlagSetUsageIncludesAutoOpen(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	fs, _ := newRuntimeFlagSet("/tmp/config-home/shortcutdeck/cards.db", &buf)
	fs.Usage()

	got := buf.String()
	if !strings.Contains(got, "-auto-open") {
		t.Fatalf("usage = %q, want auto-open flag", got)
	}
	if !strings.Contains(got, "default true") {
		t.Fatalf("usage = %q, want default true guidance", got)
	}
}

func TestNormalizeRuntimeArgsAllowsSeparateBooleanValue(t *testing.T) {
	t.Parallel()

	got := normalizeRuntimeArgs([]string{"--db", "/tmp/cards.db", "--auto-open", "false"})
	want := []string{"--db", "/tmp/cards.db", "--auto-open=false"}
	if len(got) != len(want) {
		t.Fatalf("normalizeRuntimeArgs() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("normalizeRuntimeArgs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListenAddr(t *testing.T) {
	t.Parallel()

	if got := listenAddr(7432); got != ":7432" {
		t.Fatalf("listenAddr() = %q, want %q", got, ":7432")
	}
}

func TestServerURL(t *testing.T) {
	t.Parallel()

	if got := serverURL(7432); got != "http://127.0.0.1:7432" {
		t.Fatalf("serverURL() = %q, want %q", got, "http://127.0.0.1:7432")
	}
}

func TestFormatStoreOpenErrorRewritesSQLitePathError(t *testing.T) {
	t.Parallel()

	err := formatStoreOpenError("/tmp/custom.db", errors.New("enable foreign keys: unable to open database file: out of memory (14)"))
	if !strings.Contains(err.Error(), "unable to open database") {
		t.Fatalf("error = %q, want generic open database message", err)
	}
	if !strings.Contains(err.Error(), "/tmp/custom.db") {
		t.Fatalf("error = %q, want db path in message", err)
	}
}

func TestFormatStoreOpenErrorPreservesOtherErrors(t *testing.T) {
	t.Parallel()

	rootErr := errors.New("permission denied")
	err := formatStoreOpenError("/tmp/custom.db", rootErr)
	if !strings.Contains(err.Error(), "open store: permission denied") {
		t.Fatalf("error = %q, want wrapped original error", err)
	}
}

func TestNewAppLoggerWritesWithoutPrefixNoise(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := newAppLogger(&buf)
	logger.Print("hello")

	got := buf.String()
	if !strings.Contains(got, "hello") {
		t.Fatalf("logged output = %q, want message", got)
	}
}

func TestStartupLogMessage(t *testing.T) {
	t.Parallel()

	got := startupLogMessage("http://127.0.0.1:7432", "/tmp/cards.db")
	if got != "shortcutdeck listening on http://127.0.0.1:7432 (db: /tmp/cards.db)" {
		t.Fatalf("startupLogMessage() = %q", got)
	}
}

func TestBrowserOpenWarningMessage(t *testing.T) {
	t.Parallel()

	got := browserOpenWarningMessage("http://127.0.0.1:7432", errors.New("command not found"))
	if got != "warning: could not open browser for http://127.0.0.1:7432: command not found" {
		t.Fatalf("browserOpenWarningMessage() = %q", got)
	}
}

func TestShutdownLogMessage(t *testing.T) {
	t.Parallel()

	if got := shutdownLogMessage(); got != "shortcutdeck shutting down" {
		t.Fatalf("shutdownLogMessage() = %q", got)
	}
}

func TestFatalLogMessage(t *testing.T) {
	t.Parallel()

	got := fatalLogMessage("listen and serve", errors.New("bind failed"))
	if got != "listen and serve: bind failed" {
		t.Fatalf("fatalLogMessage() = %q", got)
	}
}
