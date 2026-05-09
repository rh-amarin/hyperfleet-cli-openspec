// Package config manages the HyperFleet CLI configuration.
// Configuration is split across two YAML files:
//   - config.yaml: static settings (sections: hyperfleet, kubernetes, maestro, port-forward, database, rabbitmq, registry)
//   - state.yaml: runtime state (flat top-level keys: active-environment, cluster-id, cluster-name, nodepool-id)
//
// Environment profiles are stored at environments/<name>.yaml and deep-merged at runtime.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// defaults holds the built-in default configuration values.
var defaults = map[string]map[string]string{
	"hyperfleet": {
		"api-url":     "http://localhost:8000",
		"api-version": "v1",
		"token":       "",
		"gcp-project": "hcm-hyperfleet",
	},
	"kubernetes": {
		"context":   "",
		"namespace": "",
	},
	"maestro": {
		"consumer":              "cluster1",
		"http-endpoint":        "http://localhost:8100",
		"grpc-endpoint":        "localhost:8090",
		"namespace":            "maestro",
	},
	"port-forward": {
		"api-port":                    "8000",
		"pg-port":                     "5432",
		"maestro-http-port":           "8100",
		"maestro-http-remote-port":    "8000",
		"maestro-grpc-port":           "8090",
		"maestro-grpc-remote-port":    "8090",
	},
	"database": {
		"host":     "localhost",
		"port":     "5432",
		"name":     "hyperfleet",
		"user":     "hyperfleet",
		"password": "foobar-bizz-buzz",
	},
	"rabbitmq": {
		"host":      "localhost",
		"mgmt-port": "15672",
		"user":      "guest",
		"password":  "guest",
		"vhost":     "/",
	},
	"registry": {
		"name": "",
	},
}

// envVarMap maps environment variable names to config paths.
var envVarMap = map[string][2]string{
	"HF_API_URL":     {"hyperfleet", "api-url"},
	"HF_API_VERSION": {"hyperfleet", "api-version"},
	"HF_TOKEN":       {"hyperfleet", "token"},
	"HF_CONTEXT":     {"kubernetes", "context"},
	"HF_NAMESPACE":   {"kubernetes", "namespace"},
}

// Store manages the HyperFleet CLI configuration.
type Store struct {
	dir     string
	config  map[string]map[string]string
	state   map[string]string
	profile map[string]map[string]string // active env profile, nil if none
}

// New creates a Store that reads from the given config directory.
func New(configDir string) *Store {
	return &Store{dir: configDir}
}

// ConfigDir returns the config directory path.
func (s *Store) ConfigDir() string { return s.dir }

// Load reads config.yaml, state.yaml, and the active environment profile from disk.
// It creates the config directory and default files if they don't exist.
func (s *Store) Load() error {
	if err := os.MkdirAll(filepath.Join(s.dir, "environments"), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	s.config = deepCopyDefaults()
	s.state = map[string]string{}
	s.profile = nil

	// Load config.yaml, creating it with defaults if absent.
	cfgPath := filepath.Join(s.dir, "config.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := writeYAMLAtomic(cfgPath, mapToAny(s.config)); err != nil {
			return fmt.Errorf("create config.yaml: %w", err)
		}
	} else {
		raw, err := os.ReadFile(cfgPath)
		if err != nil {
			return fmt.Errorf("read config.yaml: %w", err)
		}
		var loaded map[string]map[string]string
		if err := yaml.Unmarshal(raw, &loaded); err != nil {
			return fmt.Errorf("parse config.yaml: %w", err)
		}
		deepMergeConfig(s.config, loaded)
	}

	// Load state.yaml, creating it empty if absent.
	statePath := filepath.Join(s.dir, "state.yaml")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		if err := writeYAMLAtomic(statePath, map[string]string{}); err != nil {
			return fmt.Errorf("create state.yaml: %w", err)
		}
	} else {
		raw, err := os.ReadFile(statePath)
		if err != nil {
			return fmt.Errorf("read state.yaml: %w", err)
		}
		if err := yaml.Unmarshal(raw, &s.state); err != nil {
			return fmt.Errorf("parse state.yaml: %w", err)
		}
		if s.state == nil {
			s.state = map[string]string{}
		}
	}

	// Load active environment profile if set.
	if active := s.state["active-environment"]; active != "" {
		profPath := filepath.Join(s.dir, "environments", active+".yaml")
		raw, err := os.ReadFile(profPath)
		if err == nil {
			var prof map[string]map[string]string
			if err := yaml.Unmarshal(raw, &prof); err == nil {
				s.profile = prof
			}
		}
	}

	return nil
}

// Get returns a configuration value using the precedence chain:
// HF_* env vars > active env profile > config.yaml > defaults.
func (s *Store) Get(section, key string) string {
	// Check environment variables first.
	envKey := envKeyFor(section, key)
	if envKey != "" {
		if v := os.Getenv(envKey); v != "" {
			return v
		}
	}

	// Active environment profile.
	if s.profile != nil {
		if sec, ok := s.profile[section]; ok {
			if v, ok := sec[key]; ok {
				return v
			}
		}
	}

	// config.yaml (already merged with defaults).
	if s.config != nil {
		if sec, ok := s.config[section]; ok {
			if v, ok := sec[key]; ok {
				return v
			}
		}
	}

	return ""
}

// Set writes a configuration value to config.yaml and updates the in-memory store.
func (s *Store) Set(section, key, value string) error {
	if s.config == nil {
		s.config = deepCopyDefaults()
	}
	if _, ok := s.config[section]; !ok {
		s.config[section] = map[string]string{}
	}
	s.config[section][key] = value
	return writeYAMLAtomic(filepath.Join(s.dir, "config.yaml"), mapToAny(s.config))
}

// GetState returns a state value from state.yaml.
func (s *Store) GetState(key string) string {
	if s.state == nil {
		return ""
	}
	return s.state[key]
}

// SetState writes a state value to state.yaml and updates in-memory state.
func (s *Store) SetState(key, value string) error {
	if s.state == nil {
		s.state = map[string]string{}
	}
	s.state[key] = value
	return writeYAMLAtomic(filepath.Join(s.dir, "state.yaml"), s.state)
}

// ActiveEnvironment returns the currently active environment name, or "" if none.
func (s *Store) ActiveEnvironment() string {
	return s.GetState("active-environment")
}

// RequireActiveEnvironment returns the active environment name or an error.
func (s *Store) RequireActiveEnvironment() (string, error) {
	env := s.ActiveEnvironment()
	if env == "" {
		return "", fmt.Errorf("[ERROR] no active environment — run 'hf config env activate <name>'")
	}
	return env, nil
}

// ClusterID resolves the cluster ID: explicit arg > state.yaml > error.
func (s *Store) ClusterID(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	id := s.GetState("cluster-id")
	if id == "" {
		return "", fmt.Errorf("[ERROR] no active cluster — run 'hf cluster use <id>' or pass --cluster-id")
	}
	return id, nil
}

// NodePoolID resolves the nodepool ID: explicit arg > state.yaml > error.
func (s *Store) NodePoolID(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	id := s.GetState("nodepool-id")
	if id == "" {
		return "", fmt.Errorf("[ERROR] no active nodepool — run 'hf nodepool use <id>' or pass --nodepool-id")
	}
	return id, nil
}

// envKeyFor returns the environment variable name for a given section/key pair.
func envKeyFor(section, key string) string {
	for envVar, path := range envVarMap {
		if path[0] == section && path[1] == key {
			return envVar
		}
	}
	return ""
}

// deepCopyDefaults returns a fresh copy of the defaults map.
func deepCopyDefaults() map[string]map[string]string {
	out := make(map[string]map[string]string, len(defaults))
	for sec, vals := range defaults {
		cp := make(map[string]string, len(vals))
		for k, v := range vals {
			cp[k] = v
		}
		out[sec] = cp
	}
	return out
}

// deepMergeConfig merges src into dst. Values in src override dst.
func deepMergeConfig(dst, src map[string]map[string]string) {
	for sec, vals := range src {
		if _, ok := dst[sec]; !ok {
			dst[sec] = map[string]string{}
		}
		for k, v := range vals {
			dst[sec][k] = v
		}
	}
}

// mapToAny converts map[string]map[string]string to map[string]any for YAML serialization.
func mapToAny(m map[string]map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for sec, vals := range m {
		inner := make(map[string]any, len(vals))
		for k, v := range vals {
			inner[k] = v
		}
		out[sec] = inner
	}
	return out
}

// writeYAMLAtomic writes data as YAML to path using a temp file + rename for atomicity.
func writeYAMLAtomic(path string, data any) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	enc := yaml.NewEncoder(tmp)
	enc.SetIndent(2)
	if err := enc.Encode(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0600); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// DefaultConfigDir returns the default config directory (~/.config/hf).
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", "hf")
	}
	return filepath.Join(home, ".config", "hf")
}

// NewFromEnv creates a Store using HF_CONFIG_DIR or the default config dir.
func NewFromEnv() *Store {
	dir := os.Getenv("HF_CONFIG_DIR")
	if dir == "" {
		dir = DefaultConfigDir()
	}
	return New(dir)
}

// ListEnvironments returns all environment names in the environments directory.
func (s *Store) ListEnvironments() ([]string, error) {
	envDir := filepath.Join(s.dir, "environments")
	entries, err := os.ReadDir(envDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return names, nil
}

// ActivateEnvironment sets the active environment in state.yaml after verifying it exists.
func (s *Store) ActivateEnvironment(name string) error {
	profPath := filepath.Join(s.dir, "environments", name+".yaml")
	if _, err := os.Stat(profPath); os.IsNotExist(err) {
		return fmt.Errorf("[ERROR] environment '%s' not found", name)
	}
	return s.SetState("active-environment", name)
}

// CountOverrides returns the number of keys overridden in an environment profile.
func (s *Store) CountOverrides(name string) (int, error) {
	profPath := filepath.Join(s.dir, "environments", name+".yaml")
	raw, err := os.ReadFile(profPath)
	if err != nil {
		return 0, err
	}
	var prof map[string]map[string]string
	if err := yaml.Unmarshal(raw, &prof); err != nil {
		return 0, err
	}
	count := 0
	for _, vals := range prof {
		count += len(vals)
	}
	return count, nil
}
