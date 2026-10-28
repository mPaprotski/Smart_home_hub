package config

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
)

func empty(string) string { return "" }
func TestDefaultsAndStrictYAML(t *testing.T) {
	for _, input := range []string{"profile: development\n", "profile: test\n"} {
		c, err := Parse([]byte(input), empty)
		if err != nil || c.Server.Host != "127.0.0.1" {
			t.Fatal(c, err)
		}
	}
	for _, input := range []string{"unknown: true", "profile: test\nprofile: production", "profile: test\n---\nprofile: test", "server:\n  port: bad", "profile: production"} {
		if _, err := Parse([]byte(input), empty); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
func TestBoundsAndMapping(t *testing.T) {
	cases := []func(*Config){func(c *Config) { c.Server.Host = "0.0.0.0" }, func(c *Config) { c.Server.Host = "192.168.1.1" }, func(c *Config) { c.Events.Queue = 0 }, func(c *Config) { c.Lighting.Zones[0].Brightness = math.NaN() }, func(c *Config) { c.Lighting.Zones[1].Offset = 0 }, func(c *Config) { c.Lighting.Zones[0].Count = 99 }, func(c *Config) { c.Lighting.Zones[0].SideCounts = []int{1, 1, 1, 0} }, func(c *Config) {
		c.Lighting.Zones[0].Kind = "matrix"
		c.Lighting.Zones[0].Width = 3
		c.Lighting.Zones[0].Height = 2
	}, func(c *Config) { c.Simulation.Tick = "0s" }, func(c *Config) { c.Components.MLEnabled = true }}
	for i, change := range cases {
		c := Defaults()
		change(&c)
		if c.Validate(false) == nil {
			t.Fatalf("case %d accepted", i)
		}
	}
	c := Defaults()
	c.Server.Host = "0.0.0.0"
	if err := c.Validate(true); err != nil {
		t.Fatal(err)
	}
}
func TestEnvironmentAndRedaction(t *testing.T) {
	secret := "never-print-this-token"
	env := func(k string) string {
		switch k {
		case "SESSION_SECRET", "ML_SERVICE_TOKEN", "MQTT_PASSWORD":
			return secret
		case "HUB_SERVER_PORT":
			return "1234"
		}
		return ""
	}
	c, err := Parse([]byte("profile: test"), env)
	if err != nil || c.Server.Port != 1234 {
		t.Fatal(err)
	}
	data, _ := json.Marshal(c)
	for _, out := range []string{string(data), fmt.Sprintf("%v", c), fmt.Sprintf("%+v", c), fmt.Sprintf("%#v", c), fmt.Sprintf("%v", c.SessionSecret)} {
		if strings.Contains(out, secret) {
			t.Fatal("secret leaked")
		}
	}
	_, err = Parse([]byte("unknown: "+secret), empty)
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatal("unsafe decode error")
	}
}

func TestSQLiteConfigurationAndAuthReadinessPolicy(t *testing.T) {
	for _, input := range []string{"storage: sqlite\nsqlite:\n  path: .runtime/data/hub.db\n  busy_timeout: 500ms\n  startup_timeout: 10s\nauth:\n  required: true", "storage: sqlite"} {
		c, err := Parse([]byte(input), empty)
		if err != nil || c.Storage != "sqlite" {
			t.Fatal(c, err)
		}
	}
	for _, input := range []string{"storage: sqlite\nsqlite:\n  path: ':memory:'", "storage: sqlite\nsqlite:\n  busy_timeout: 60s", "auth:\n  required: true"} {
		if _, err := Parse([]byte(input), empty); err == nil {
			t.Fatal(input)
		}
	}
	c, err := Parse([]byte("storage: sqlite"), func(k string) string {
		if k == "HUB_DB_PATH" {
			return "private/data/hub.db"
		}
		return ""
	})
	if err != nil || c.SQLite.Path != "private/data/hub.db" {
		t.Fatal(err)
	}
	if strings.Contains(c.String(), "private/data") {
		t.Fatal("DB path leaked")
	}
}
