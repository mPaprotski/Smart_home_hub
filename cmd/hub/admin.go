package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"golang.org/x/term"
	"io"
	"os"
	"time"
)

func openStore(ctx context.Context, c config.Config) (*storage.SQLite, error) {
	busy, _ := time.ParseDuration(c.SQLite.BusyTimeout)
	startup, _ := time.ParseDuration(c.SQLite.StartupTimeout)
	return storage.Open(ctx, storage.Options{Path: c.SQLite.Path, BusyTimeout: busy, StartupTimeout: startup})
}
func admin(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: hub admin bootstrap|verify --config path --username name [--password-stdin]")
		return 2
	}
	action := args[0]
	fs := flag.NewFlagSet("admin", flag.ContinueOnError)
	path := fs.String("config", "config.example.yaml", "configuration path")
	username := fs.String("username", "", "account name")
	stdin := fs.Bool("password-stdin", false, "explicit bounded stdin input (CI synthetic credentials)")
	if fs.Parse(args[1:]) != nil || fs.NArg() != 0 || (action != "bootstrap" && action != "verify") {
		return 2
	}
	cfg, err := config.Load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if cfg.Storage != "sqlite" {
		fmt.Fprintln(os.Stderr, "BOOTSTRAP_REQUIRES_SQLITE")
		return 2
	}
	password, err := readPassword(*stdin, action == "bootstrap")
	if err != nil {
		fmt.Fprintln(os.Stderr, "PASSWORD_INPUT_REJECTED")
		return 2
	}
	defer clear(password)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := openStore(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer store.Close()
	service, err := auth.New(ctx, store, auth.NewHasher())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if action == "bootstrap" {
		_, err = service.Bootstrap(ctx, *username, password)
	} else {
		_, err = service.Authenticate(ctx, *username, password, "local-cli")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	message := "administrator_created"
	if action == "verify" {
		message = "credentials_verified_no_session"
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]string{"result": message})
	return 0
}
func readPassword(stdin, confirmation bool) ([]byte, error) {
	if stdin {
		data, err := io.ReadAll(io.LimitReader(os.Stdin, 1027))
		if err != nil || len(data) > 1026 {
			return nil, auth.ErrPassword
		}
		if bytes.HasSuffix(data, []byte("\r\n")) {
			data = data[:len(data)-2]
		} else {
			data = bytes.TrimSuffix(data, []byte("\n"))
		}
		if len(data) > 1024 {
			return nil, auth.ErrPassword
		}
		return data, nil
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, auth.ErrPassword
	}
	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, err
	}
	if confirmation {
		fmt.Fprint(os.Stderr, "Confirm: ")
		second, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		defer clear(second)
		if err != nil || !bytes.Equal(password, second) {
			clear(password)
			return nil, auth.ErrPassword
		}
	}
	return password, nil
}
