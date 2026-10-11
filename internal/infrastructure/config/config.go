package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration for the worker.
type Config struct {
	API struct {
		BaseURL string
		APIKey  string
		Timeout time.Duration
	}
	Tailers struct {
		Files        []string
		JournaldUnit string
	}
	Worker struct {
		BatchSize     int
		BatchInterval time.Duration
		Concurrency   int
	}
	Spool struct {
		Directory     string
		FlushInterval time.Duration
	}
	Metrics struct {
		ListenAddr string
	}
	HostID string
	DryRun bool
}

// Load reads configuration from YAML (limited subset) and environment variables.
func Load(path string) (Config, error) {
	cfg := Config{}
	if path != "" {
		if err := parseFile(path, &cfg); err != nil {
			return Config{}, err
		}
	}

	applyEnvOverrides(&cfg)
	applyDerivedDefaults(&cfg)
	return cfg, nil
}

// LoadFromEnv builds a configuration purely from environment variables.
func LoadFromEnv() (Config, error) {
	cfg := Config{}
	applyEnvOverrides(&cfg)
	applyDerivedDefaults(&cfg)
	return cfg, nil
}

func parseFile(path string, cfg *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var section string
	var listTarget *[]string

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := leadingSpaces(raw)

		if indent == 0 {
			if !strings.HasSuffix(trimmed, ":") {
				return fmt.Errorf("invalid section header on line %d", lineNo)
			}
			section = strings.TrimSuffix(trimmed, ":")
			listTarget = nil
			continue
		}

		if indent == 2 {
			if strings.HasSuffix(trimmed, ":") {
				key := strings.TrimSuffix(trimmed, ":")
				if section == "tailers" && key == "files" {
					cfg.Tailers.Files = cfg.Tailers.Files[:0]
					listTarget = &cfg.Tailers.Files
				}
				continue
			}

			key, value, err := splitKeyValue(trimmed)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			if err := assignValue(cfg, section, key, value); err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			continue
		}

		if indent == 4 && strings.HasPrefix(strings.TrimSpace(raw), "- ") {
			val := strings.TrimSpace(raw)[2:]
			val = strings.Trim(val, "\"'")
			val = os.ExpandEnv(val)
			if listTarget != nil {
				*listTarget = append(*listTarget, val)
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func splitKeyValue(line string) (string, string, error) {
	idx := strings.Index(line, ":")
	if idx == -1 {
		return "", "", errors.New("expected key: value pair")
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	value = strings.Trim(value, "\"'")
	value = os.ExpandEnv(value)
	return key, value, nil
}

func assignValue(cfg *Config, section, key, value string) error {
	switch section {
	case "api":
		switch key {
		case "base_url":
			cfg.API.BaseURL = value
		case "api_key":
			cfg.API.APIKey = value
		case "timeout":
			if value == "" {
				return nil
			}
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration for api.timeout: %w", err)
			}
			cfg.API.Timeout = d
		}
	case "tailers":
		switch key {
		case "journald_unit":
			cfg.Tailers.JournaldUnit = value
		}
	case "worker":
		switch key {
		case "batch_size":
			i, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid worker.batch_size: %w", err)
			}
			cfg.Worker.BatchSize = i
		case "batch_interval":
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid worker.batch_interval: %w", err)
			}
			cfg.Worker.BatchInterval = d
		case "concurrency":
			i, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid worker.concurrency: %w", err)
			}
			cfg.Worker.Concurrency = i
		}
	case "spool":
		switch key {
		case "directory":
			cfg.Spool.Directory = value
		case "flush_interval":
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid spool.flush_interval: %w", err)
			}
			cfg.Spool.FlushInterval = d
		}
	case "metrics":
		if key == "listen_addr" {
			cfg.Metrics.ListenAddr = value
		}
	case "":
		return fmt.Errorf("value outside of section: %s", key)
	default:
		if key == "host_id" {
			cfg.HostID = value
		} else if key == "dry_run" {
			cfg.DryRun = parseBool(value)
		}
	}
	return nil
}

func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func applyEnvOverrides(cfg *Config) {
	setStringEnv(&cfg.API.BaseURL, "MINI_SIEM_API_BASE_URL")
	setStringEnv(&cfg.API.APIKey, "MINI_SIEM_API_KEY")
	setDurationEnv(&cfg.API.Timeout, "MINI_SIEM_API_TIMEOUT")

	setStringEnv(&cfg.Spool.Directory, "MINI_SIEM_SPOOL_DIR")
	setDurationEnv(&cfg.Spool.FlushInterval, "MINI_SIEM_SPOOL_FLUSH_INTERVAL")

	setIntEnv(&cfg.Worker.BatchSize, "MINI_SIEM_WORKER_BATCH_SIZE")
	setDurationEnv(&cfg.Worker.BatchInterval, "MINI_SIEM_WORKER_BATCH_INTERVAL")
	setIntEnv(&cfg.Worker.Concurrency, "MINI_SIEM_WORKER_CONCURRENCY")

	setStringEnv(&cfg.Metrics.ListenAddr, "MINI_SIEM_METRICS_ADDR")
	setStringEnv(&cfg.HostID, "MINI_SIEM_HOST_ID")

	if v, ok := os.LookupEnv("MINI_SIEM_DRY_RUN"); ok {
		cfg.DryRun = parseBool(v)
	}

	if v, ok := os.LookupEnv("MINI_SIEM_TAILER_JOURNALD_UNIT"); ok {
		cfg.Tailers.JournaldUnit = v
	}
	if v, ok := os.LookupEnv("MINI_SIEM_TAILER_FILES"); ok {
		cfg.Tailers.Files = splitList(v)
	}
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func setStringEnv(target *string, key string) {
	if v, ok := os.LookupEnv(key); ok {
		*target = v
	}
}

func setIntEnv(target *int, key string) {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			*target = i
		}
	}
}

func setDurationEnv(target *time.Duration, key string) {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			*target = d
		}
	}
}

func applyDerivedDefaults(cfg *Config) {
	if cfg.API.Timeout == 0 {
		cfg.API.Timeout = 10 * time.Second
	}
	if cfg.Worker.BatchSize <= 0 {
		cfg.Worker.BatchSize = 100
	}
	if cfg.Worker.BatchInterval <= 0 {
		cfg.Worker.BatchInterval = 2 * time.Second
	}
	if cfg.Worker.Concurrency <= 0 {
		cfg.Worker.Concurrency = 4
	}
	if cfg.Spool.Directory == "" {
		cfg.Spool.Directory = "/var/lib/mini-siem/spool"
	}
	if cfg.Spool.FlushInterval <= 0 {
		cfg.Spool.FlushInterval = 30 * time.Second
	}
	if cfg.Metrics.ListenAddr == "" {
		cfg.Metrics.ListenAddr = ":2112"
	}
	if cfg.HostID == "" {
		if hostname, err := os.Hostname(); err == nil {
			cfg.HostID = hostname
		}
	}
}

func leadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
			continue
		}
		break
	}
	return count
}
