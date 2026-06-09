package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Environment string

const (
	EnvDev   Environment = "dev"
	EnvStage Environment = "stage"
	EnvProd  Environment = "prod"
)

func AppEnv() Environment {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "stage", "staging":
		return EnvStage
	case "prod", "production":
		return EnvProd
	default:
		return EnvDev
	}
}

func IsProd() bool  { return AppEnv() == EnvProd }
func IsStage() bool { return AppEnv() == EnvStage }
func IsDev() bool   { return AppEnv() == EnvDev }

func requireInNonDev(key string) {
	if (IsProd() || IsStage()) && strings.TrimSpace(os.Getenv(key)) == "" {
		log.Fatalf("required environment variable %s is not set (APP_ENV=%s)", key, AppEnv())
	}
}

// String returns env value or devDefault in dev; required in stage/prod.
func String(key, devDefault string) string {
	requireInNonDev(key)
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return devDefault
}

// MustString always requires the variable.
func MustString(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func Int(key string, devDefault int) int {
	requireInNonDev(key)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return devDefault
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		log.Fatalf("invalid integer for %s: %q", key, raw)
	}
	return n
}

func Float(key string, devDefault float64) float64 {
	requireInNonDev(key)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return devDefault
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		log.Fatalf("invalid float for %s: %q", key, raw)
	}
	return f
}

func Bool(key string, devDefault bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return devDefault
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		log.Fatalf("invalid bool for %s: %q", key, raw)
	}
	return b
}

func Duration(key string, devDefault time.Duration) time.Duration {
	requireInNonDev(key)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return devDefault
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Fatalf("invalid duration for %s: %q", key, raw)
	}
	return d
}

func CSV(key string, devDefault []string) []string {
	requireInNonDev(key)
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return devDefault
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func KafkaBrokers() []string {
	return CSV("KAFKA_BROKERS", []string{"localhost:9092"})
}

func CORSOrigins(devAllowAll bool) []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		if devAllowAll && IsDev() {
			return []string{"*"}
		}
		return nil
	}
	if raw == "*" {
		return []string{"*"}
	}
	return CSV("CORS_ALLOWED_ORIGINS", nil)
}
