// Package config owns infrastructure configuration; runtime user settings belong to Storage.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.yaml.in/yaml/v3"
)

type Error struct{ Code, Field string }

func (e *Error) Error() string { return e.Code + ": " + e.Field }

type Secret struct{ value string }

func (s Secret) Empty() bool                  { return s.value == "" }
func (s Secret) String() string               { return "[REDACTED]" }
func (s Secret) GoString() string             { return s.String() }
func (s Secret) LogValue() slog.Value         { return slog.StringValue(s.String()) }
func (s Secret) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

type Server struct {
	TLSCert           string `yaml:"tls_cert" json:"-"`
	TLSKey            string `yaml:"tls_key" json:"-"`
	Host              string `yaml:"host" json:"host"`
	Port              int    `yaml:"port" json:"port"`
	ReadHeaderTimeout string `yaml:"read_header_timeout" json:"read_header_timeout"`
	ShutdownTimeout   string `yaml:"shutdown_timeout" json:"shutdown_timeout"`
}
type Events struct {
	Queue            int `yaml:"queue" json:"queue"`
	MaxSubscriptions int `yaml:"max_subscriptions" json:"max_subscriptions"`
	MaxPayloadBytes  int `yaml:"max_payload_bytes" json:"max_payload_bytes"`
}
type Zone struct {
	ID            string  `yaml:"id" json:"id"`
	Offset        int     `yaml:"offset" json:"offset"`
	Count         int     `yaml:"count" json:"count"`
	Kind          string  `yaml:"kind" json:"kind"`
	Width         int     `yaml:"width" json:"width"`
	Height        int     `yaml:"height" json:"height"`
	SideCounts    []int   `yaml:"side_counts,omitempty" json:"side_counts,omitempty"`
	Brightness    float64 `yaml:"brightness" json:"brightness"`
	MaxCurrentAmp float64 `yaml:"max_current_amp" json:"max_current_amp"`
}
type Lighting struct {
	Capacity            int     `yaml:"capacity" json:"capacity"`
	GlobalMaxCurrentAmp float64 `yaml:"global_max_current_amp" json:"global_max_current_amp"`
	Zones               []Zone  `yaml:"zones" json:"zones"`
}
type Simulation struct {
	Seed int64  `yaml:"seed" json:"seed"`
	Tick string `yaml:"tick" json:"tick"`
}
type Components struct {
	MLEnabled   bool `yaml:"ml_enabled" json:"ml_enabled"`
	MQTTEnabled bool `yaml:"mqtt_enabled" json:"mqtt_enabled"`
}
type SQLite struct {
	Path           string `yaml:"path" json:"-"`
	BusyTimeout    string `yaml:"busy_timeout" json:"busy_timeout"`
	StartupTimeout string `yaml:"startup_timeout" json:"startup_timeout"`
}
type Auth struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	PublicOrigin   string   `yaml:"public_origin" json:"public_origin"`
	TrustedProxies []string `yaml:"trusted_proxies" json:"trusted_proxies"`
	Required       bool     `yaml:"required" json:"required"`
}
type Config struct {
	SQLite        SQLite     `yaml:"sqlite" json:"sqlite"`
	Auth          Auth       `yaml:"auth" json:"auth"`
	Profile       string     `yaml:"profile" json:"profile"`
	AdapterMode   string     `yaml:"adapter_mode" json:"adapter_mode"`
	Storage       string     `yaml:"storage" json:"storage"`
	LogLevel      string     `yaml:"log_level" json:"log_level"`
	Server        Server     `yaml:"server" json:"server"`
	Events        Events     `yaml:"events" json:"events"`
	Lighting      Lighting   `yaml:"lighting" json:"lighting"`
	Simulation    Simulation `yaml:"simulation" json:"simulation"`
	Components    Components `yaml:"components" json:"components"`
	SessionSecret Secret     `yaml:"-" json:"-"`
	MLToken       Secret     `yaml:"-" json:"-"`
	MQTTPassword  Secret     `yaml:"-" json:"-"`
}

func (c Config) String() string       { b, _ := json.Marshal(c); return string(b) }
func (c Config) GoString() string     { return c.String() }
func (c Config) LogValue() slog.Value { return slog.StringValue(c.String()) }
func Defaults() Config {
	return Config{SQLite: SQLite{Path: ".runtime/data/hub.db", BusyTimeout: "500ms", StartupTimeout: "10s"}, Profile: "development", AdapterMode: "simulation", Storage: "memory_fake", LogLevel: "INFO", Server: Server{Host: "127.0.0.1", Port: 8080, ReadHeaderTimeout: "5s", ShutdownTimeout: "5s"}, Events: Events{Queue: 64, MaxSubscriptions: 128, MaxPayloadBytes: 16384}, Lighting: Lighting{Capacity: 8, GlobalMaxCurrentAmp: 1, Zones: []Zone{{ID: "monitor.backlight", Count: 4, Kind: "strip", Brightness: .2, MaxCurrentAmp: .5, SideCounts: []int{1, 1, 1, 1}}, {ID: "desk.wall", Offset: 4, Count: 4, Kind: "strip", Brightness: .2, MaxCurrentAmp: .5}}}, Simulation: Simulation{Seed: 42, Tick: "100ms"}}
}
func Parse(data []byte, getenv func(string) string) (Config, error) {
	c := Defaults()
	if len(data) > 1<<20 {
		return c, &Error{"CONFIG_SIZE", "yaml"}
	}
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&c); err != nil {
		return c, &Error{"CONFIG_PARSE", "yaml"}
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return c, &Error{"CONFIG_PARSE", "multiple documents"}
	}
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	for key, dst := range map[string]*string{"HUB_PROFILE": &c.Profile, "HUB_ADAPTER_MODE": &c.AdapterMode, "HUB_STORAGE": &c.Storage, "HUB_SERVER_HOST": &c.Server.Host, "HUB_LOG_LEVEL": &c.LogLevel, "HUB_DB_PATH": &c.SQLite.Path, "HUB_TLS_CERT": &c.Server.TLSCert, "HUB_TLS_KEY": &c.Server.TLSKey, "HUB_PUBLIC_ORIGIN": &c.Auth.PublicOrigin} {
		if value := getenv(key); value != "" {
			*dst = value
		}
	}
	if value := getenv("HUB_SERVER_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return c, &Error{"CONFIG_VALUE", "server.port"}
		}
		c.Server.Port = port
	}
	c.SessionSecret = Secret{getenv("SESSION_SECRET")}
	c.MLToken = Secret{getenv("ML_SERVICE_TOKEN")}
	c.MQTTPassword = Secret{getenv("MQTT_PASSWORD")}
	return c, c.Validate(getenv("HUB_CONTAINER_NETWORK") == "true")
}
func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, &Error{"CONFIG_READ", "path"}
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return Config{}, &Error{"CONFIG_READ", "path"}
	}
	return Parse(data, os.Getenv)
}
func finitePositive(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) && x > 0 }
func (c Config) Validate(containerNetwork bool) error {
	fail := func(field string) error { return &Error{"CONFIG_VALUE", field} }
	if c.Profile != "development" && c.Profile != "test" && c.Profile != "production" {
		return fail("profile")
	}
	if c.Profile == "production" {
		return &Error{"NOT_IMPLEMENTED", "production auth/storage/hardware"}
	}
	if c.AdapterMode != "simulation" {
		return &Error{"NOT_IMPLEMENTED", "hardware adapters"}
	}
	if c.Storage != "memory_fake" && c.Storage != "not_implemented" && c.Storage != "sqlite" {
		return fail("storage")
	}
	if c.Storage == "sqlite" {
		if c.SQLite.Path == "" || c.SQLite.Path == ":memory:" || len(c.SQLite.Path) > 4096 || filepath.Clean(c.SQLite.Path) == "." {
			return fail("sqlite.path")
		}
		for name, value := range map[string]string{"sqlite.busy_timeout": c.SQLite.BusyTimeout, "sqlite.startup_timeout": c.SQLite.StartupTimeout} {
			d, e := time.ParseDuration(value)
			maximum := time.Minute
			if name == "sqlite.busy_timeout" {
				maximum = 5 * time.Second
			}
			if e != nil || d < time.Millisecond || d > maximum {
				return fail(name)
			}
		}
	}
	if (c.Server.TLSCert == "") != (c.Server.TLSKey == "") {
		return fail("server.tls")
	}
	if c.Auth.Enabled {
		if c.Storage != "sqlite" {
			return fail("auth requires sqlite")
		}
		u, err := url.Parse(c.Auth.PublicOrigin)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return fail("auth.public_origin")
		}
		if len(c.Auth.TrustedProxies) > 8 {
			return fail("auth.trusted_proxies")
		}
		for _, cidr := range c.Auth.TrustedProxies {
			_, network, e := net.ParseCIDR(cidr)
			if e != nil {
				return fail("auth.trusted_proxies")
			}
			ones, bits := network.Mask.Size()
			if ones != bits {
				return fail("auth.trusted_proxies")
			}
		}
		if c.Server.TLSCert == "" && len(c.Auth.TrustedProxies) == 0 {
			return fail("auth requires TLS or explicit trusted proxy")
		}
	}
	if c.Auth.Required && c.Storage != "sqlite" {
		return fail("auth requires sqlite")
	}
	if ip := net.ParseIP(c.Server.Host); ip == nil || (!ip.IsLoopback() && !(containerNetwork && c.Server.Host == "0.0.0.0")) {
		return fail("server.host")
	}
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fail("server.port")
	}
	for field, value := range map[string]string{"server.read_header_timeout": c.Server.ReadHeaderTimeout, "server.shutdown_timeout": c.Server.ShutdownTimeout, "simulation.tick": c.Simulation.Tick} {
		d, e := time.ParseDuration(value)
		if e != nil || d <= 0 || d > time.Minute {
			return fail(field)
		}
	}
	if d, _ := time.ParseDuration(c.Simulation.Tick); d < time.Millisecond {
		return fail("simulation.tick")
	}
	if c.Events.Queue < 1 || c.Events.Queue > 4096 || c.Events.MaxSubscriptions < 1 || c.Events.MaxSubscriptions > 4096 || c.Events.MaxPayloadBytes < 128 || c.Events.MaxPayloadBytes > 1<<20 {
		return fail("events limits")
	}
	if c.LogLevel != "DEBUG" && c.LogLevel != "INFO" && c.LogLevel != "WARN" && c.LogLevel != "ERROR" {
		return fail("log_level")
	}
	if c.Lighting.Capacity < 1 || c.Lighting.Capacity > 4096 || !finitePositive(c.Lighting.GlobalMaxCurrentAmp) || len(c.Lighting.Zones) == 0 {
		return fail("lighting")
	}
	used := make([]bool, c.Lighting.Capacity)
	ids := map[string]bool{}
	for _, z := range c.Lighting.Zones {
		if z.ID == "" || ids[z.ID] || z.Offset < 0 || z.Count < 1 || z.Offset > len(used) || z.Count > len(used)-z.Offset || !finitePositive(z.MaxCurrentAmp) || math.IsNaN(z.Brightness) || math.IsInf(z.Brightness, 0) || z.Brightness < 0 || z.Brightness > 1 {
			return fail("lighting.zone")
		}
		ids[z.ID] = true
		if z.Kind != "strip" && z.Kind != "matrix" {
			return fail("lighting.zone.kind")
		}
		if z.Kind == "matrix" && (z.Width < 1 || z.Height < 1 || z.Width > 4096 || z.Height > 4096 || z.Width*z.Height != z.Count) {
			return fail("lighting.matrix")
		}
		if len(z.SideCounts) > 0 {
			sum := 0
			for _, v := range z.SideCounts {
				if v < 0 || v > 4096 {
					return fail("lighting.sides")
				}
				sum += v
			}
			if len(z.SideCounts) != 4 || sum != z.Count {
				return fail("lighting.sides")
			}
		}
		for i := z.Offset; i < z.Offset+z.Count; i++ {
			if used[i] {
				return fail("lighting.overlap")
			}
			used[i] = true
		}
	}
	if c.Components.MLEnabled && c.MLToken.Empty() {
		return fail("ML_SERVICE_TOKEN required")
	}
	if c.Components.MQTTEnabled && c.MQTTPassword.Empty() {
		return fail("MQTT_PASSWORD required")
	}
	return nil
}
func (c Config) Address() string { return net.JoinHostPort(c.Server.Host, strconv.Itoa(c.Server.Port)) }
func (c Config) ShutdownDuration() time.Duration {
	d, _ := time.ParseDuration(c.Server.ShutdownTimeout)
	return d
}
func (c Config) TickDuration() time.Duration { d, _ := time.ParseDuration(c.Simulation.Tick); return d }
func NewError(code, field string) error      { return &Error{code, field} }

var _ fmt.Stringer = Config{}
