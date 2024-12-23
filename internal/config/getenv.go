package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"effective-mobile-go/internal/lib"
)

type getenv struct {
	err error
}

var ErrEnvRequired = errors.New("env is required")

func (ge *getenv) String(key string, required bool, defaultValue string) string {

	if ge.err != nil {
		return ""
	}

	if s, ok := os.LookupEnv(key); ok {
		return s
	}

	if required {
		ge.err = fmt.Errorf("%s %w", key, ErrEnvRequired)
		return ""
	}

	return defaultValue
}

func (ge *getenv) Int(key string, required bool, defaultValue int) int {

	if ge.err != nil {
		return 0
	}

	if s, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(s)
		if err != nil {
			ge.err = err
			return 0
		}
		return v
	}

	if required {
		ge.err = fmt.Errorf("%s %w", key, ErrEnvRequired)
		return 0
	}

	return defaultValue
}

func (ge *getenv) LogLevel(key string, required bool, defaultValue slog.Level) slog.Level {

	if ge.err != nil {
		return 0
	}

	if s, ok := os.LookupEnv(key); ok {
		var v slog.Level
		if err := v.UnmarshalText(lib.StringAsBytes(s)); err != nil {
			ge.err = err
			return 0
		}
		return v
	}

	if required {
		ge.err = fmt.Errorf("%s %w", key, ErrEnvRequired)
		return 0
	}

	return defaultValue
}

func (ge *getenv) Bool(key string, required bool, defaultValue bool) bool {

	if ge.err != nil {
		return false
	}

	if s, ok := os.LookupEnv(key); ok {

		switch strings.ToLower(s) {
		case "true", "yes", "on", "1":
			return true
		case "false", "no", "off", "0":
			return false
		default:
			msg := fmt.Sprintf("%s=%s env is ignored. Want value: true/false, yes/no, on/off or 1/0", key, s)
			if required {
				ge.err = errors.New(msg)
			} else {
				slog.Error(msg)
			}
			return false
		}

	}

	if required {
		ge.err = fmt.Errorf("%s %w", key, ErrEnvRequired)
		return false
	}

	return defaultValue
}
