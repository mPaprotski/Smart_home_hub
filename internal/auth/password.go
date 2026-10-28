package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalid     = errors.New("INVALID_CREDENTIALS")
	ErrRate        = errors.New("RATE_LIMITED")
	ErrPassword    = errors.New("PASSWORD_POLICY")
	ErrUnavailable = errors.New("AUTH_UNAVAILABLE")
)

const (
	MemoryKiB   uint32 = 19456
	Iterations  uint32 = 2
	Parallelism uint8  = 1
)

type Hasher struct {
	slots  chan struct{}
	random io.Reader
}

func NewHasher() *Hasher { return &Hasher{slots: make(chan struct{}, 1), random: rand.Reader} }
func PasswordPolicy(password []byte) error {
	if len(password) > 1024 || !utf8.Valid(password) || utf8.RuneCount(password) < 15 {
		return ErrPassword
	}
	return nil
}
func (h *Hasher) enter(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrUnavailable
	}
	select {
	case h.slots <- struct{}{}:
		return nil
	default:
		return ErrRate
	}
}
func (h *Hasher) Hash(ctx context.Context, password []byte) (string, error) {
	if PasswordPolicy(password) != nil {
		return "", ErrPassword
	}
	if err := h.enter(ctx); err != nil {
		return "", err
	}
	defer func() { <-h.slots }()
	salt := make([]byte, 16)
	if _, err := io.ReadFull(h.random, salt); err != nil {
		return "", ErrUnavailable
	}
	key := argon2.IDKey(password, salt, Iterations, MemoryKiB, Parallelism, 32)
	if ctx.Err() != nil {
		return "", ErrUnavailable
	}
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", MemoryKiB, Iterations, Parallelism, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

type parameters struct {
	memory, iterations uint32
	parallelism        uint8
	salt, key          []byte
}

func parseHash(encoded string) (parameters, error) {
	var p parameters
	if len(encoded) > 256 {
		return p, ErrInvalid
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return p, ErrInvalid
	}
	fields := strings.Split(parts[3], ",")
	if len(fields) != 3 {
		return p, ErrInvalid
	}
	values := make([]uint64, 3)
	for i, prefix := range []string{"m=", "t=", "p="} {
		if !strings.HasPrefix(fields[i], prefix) {
			return p, ErrInvalid
		}
		v, err := strconv.ParseUint(strings.TrimPrefix(fields[i], prefix), 10, 32)
		if err != nil {
			return p, ErrInvalid
		}
		values[i] = v
	}
	if values[0] < 19456 || values[0] > 65536 || values[1] < 2 || values[1] > 4 || values[2] < 1 || values[2] > 4 {
		return p, ErrInvalid
	}
	p.memory = uint32(values[0])
	p.iterations = uint32(values[1])
	p.parallelism = uint8(values[2])
	var err error
	p.salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(p.salt) != 16 {
		return p, ErrInvalid
	}
	p.key, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(p.key) != 32 {
		return p, ErrInvalid
	}
	return p, nil
}
func (h *Hasher) Verify(ctx context.Context, encoded string, password []byte) error {
	if len(password) == 0 || len(password) > 1024 || !utf8.Valid(password) {
		return ErrInvalid
	}
	p, err := parseHash(encoded)
	if err != nil {
		return err
	}
	if err = h.enter(ctx); err != nil {
		return err
	}
	defer func() { <-h.slots }()
	got := argon2.IDKey(password, p.salt, p.iterations, p.memory, p.parallelism, 32)
	if ctx.Err() != nil {
		return ErrUnavailable
	}
	if subtle.ConstantTimeCompare(got, p.key) != 1 {
		return ErrInvalid
	}
	return nil
}
func NeedsRehash(encoded string) bool {
	p, err := parseHash(encoded)
	return err != nil || p.memory < MemoryKiB || p.iterations < Iterations || p.parallelism < Parallelism
}
