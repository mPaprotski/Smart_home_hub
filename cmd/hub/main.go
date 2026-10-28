package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/api"
	"github.com/mPaprotski/Smart_home_hub/internal/app"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/buildinfo"
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"github.com/mPaprotski/Smart_home_hub/internal/logging"
	"github.com/mPaprotski/Smart_home_hub/internal/simulation"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"github.com/mPaprotski/Smart_home_hub/internal/system"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "admin" {
		return admin(os.Args[2:])
	}
	path := flag.String("config", "config.example.yaml", "explicit YAML configuration path")
	version := flag.Bool("version", false, "print build version")
	check := flag.Bool("check-config", false, "validate configuration without opening devices or sockets")
	probe := flag.String("healthcheck", "", "probe a local readiness URL and exit")
	probeCA := flag.String("healthcheck-ca", "", "explicit PEM CA file for HTTPS readiness probe")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		return 2
	}
	if *version {
		fmt.Println(buildinfo.Version, buildinfo.Commit)
		return 0
	}
	if *probe != "" {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		if *probeCA != "" {
			data, err := os.ReadFile(*probeCA)
			if err != nil || len(data) > 1<<20 {
				return 1
			}
			roots := x509.NewCertPool()
			if !roots.AppendCertsFromPEM(data) {
				return 1
			}
			transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
		}
		client := http.Client{Transport: transport, Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := client.Get(*probe)
		if err != nil {
			return 1
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return 1
		}
		return 0
	}
	cfg, err := config.Load(*path)
	logger := logging.New(os.Stderr, cfg.LogLevel)
	if err != nil {
		logger.Error("configuration_rejected", "error", err.Error())
		return 2
	}
	if *check {
		logger.Info("configuration_valid", "config", cfg)
		return 0
	}
	var boot [16]byte
	if _, err := rand.Read(boot[:]); err != nil {
		logger.Error("boot_id_failed")
		return 1
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	bus := events.New(cfg.Events.MaxSubscriptions, cfg.Events.MaxPayloadBytes)
	defer bus.Close()
	fatal := make(chan error, 1)
	health := system.New(cfg)
	observer := simulation.NewObserver(bus, storage.NewMemory(5), cfg.Events.Queue, fatal)
	suite := simulation.New(bus, cfg.Simulation.Seed, cfg.TickDuration(), hex.EncodeToString(boot[:]), fatal)
	server := api.New(cfg, health, bus, logger, fatal)
	// Health is wired after successful fake initialization, never inferred from config alone.
	adapters := &adapterComponent{suite: suite, health: health}
	components := []app.Component{}
	if cfg.Storage == "sqlite" {
		persistent := &persistentComponent{cfg: cfg, health: health, server: server, fatal: fatal}
		components = append(components, persistent)
	}
	components = append(components, observer, adapters, server)
	err = app.Run(ctx, cfg.ShutdownDuration(), fatal, components...)
	if err != nil {
		logger.Error("hub_failed", "error", err.Error())
		return 1
	}
	logger.Info("hub_stopped", "observations", observer.Count())
	return 0
}

type adapterComponent struct {
	suite  *simulation.Suite
	health *system.Health
}

func (a *adapterComponent) Start(ctx context.Context) error {
	if err := a.suite.Start(ctx); err != nil {
		return err
	}
	a.health.Set("adapters", "READY")
	a.health.SetPhase("RUNNING")
	return nil
}
func (a *adapterComponent) Stop(ctx context.Context) error {
	a.health.Set("adapters", "UNAVAILABLE")
	return a.suite.Stop(ctx)
}
func main() { os.Exit(run()) }

type persistentComponent struct {
	cfg    config.Config
	health *system.Health
	server *api.Server
	store  *storage.SQLite
	cancel context.CancelFunc
	done   chan struct{}
	fatal  chan<- error
}

func (p *persistentComponent) Start(ctx context.Context) error {
	store, err := openStore(ctx, p.cfg)
	if err != nil {
		return err
	}
	service, err := auth.New(ctx, store, auth.NewHasher())
	if err != nil {
		_ = store.Close()
		return err
	}
	if _, err = store.Cleanup(ctx, time.Now().UTC(), 100); err != nil {
		_ = store.Close()
		return err
	}
	p.store = store
	p.server.SetAuthenticator(service)
	if p.cfg.Auth.Enabled {
		p.server.SetAccess(auth.NewAccess(store, service))
		p.health.Set("auth", "READY")
	}
	p.health.Set("storage", "READY")
	child, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.done = make(chan struct{})
	go func() {
		defer close(p.done)
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-child.Done():
				return
			case <-ticker.C:
				cleanup, done := context.WithTimeout(child, 5*time.Second)
				_, err := store.Cleanup(cleanup, time.Now().UTC(), 100)
				done()
				if err != nil && child.Err() == nil {
					p.health.Set("storage", "UNAVAILABLE")
					select {
					case p.fatal <- errors.New("storage maintenance failed"):
					case <-child.Done():
					}
					return
				}
			}
		}
	}()
	return nil
}
func (p *persistentComponent) Stop(ctx context.Context) error {
	if p.cancel != nil {
		p.cancel()
		select {
		case <-p.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	p.health.Set("storage", "UNAVAILABLE")
	return p.store.Close()
}
