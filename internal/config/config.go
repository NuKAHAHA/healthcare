package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Logger    LoggerConfig
	RateLimit RateLimitConfig
	CORS      CORSConfig
	AppEnv    string
}

type ServerConfig struct {
	Port           int
	MaxBodySize    int64
	RequestTimeout int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret           string
	AccessExpireMin  int
	RefreshExpireDay int
}

type LoggerConfig struct {
	Level string
}

type RateLimitConfig struct {
	GlobalRPM        int
	LoginMaxAttempts int
	LoginWindowMin   int
	LoginBlockMin    int
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	loadDotEnv()

	rawOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	origins := splitTrim(rawOrigins, ",")

	cfg := &Config{
		AppEnv: getEnv("APP_ENV", "production"),
		Server: ServerConfig{
			Port:           getEnvInt("SERVER_PORT", 8080),
			MaxBodySize:    getEnvInt64("MAX_BODY_SIZE", 1048576),
			RequestTimeout: getEnvInt("REQUEST_TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "healthcare"),
			// Default to require; override with DB_SSL_MODE=disable for local dev
			SSLMode: getEnv("DB_SSL_MODE", "require"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", ""),
			AccessExpireMin:  getEnvInt("JWT_ACCESS_EXPIRE_MIN", 5),
			RefreshExpireDay: getEnvInt("JWT_REFRESH_EXPIRE_DAY", 7),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		RateLimit: RateLimitConfig{
			GlobalRPM:        getEnvInt("RATE_LIMIT_GLOBAL_RPM", 100),
			LoginMaxAttempts: getEnvInt("LOGIN_MAX_ATTEMPTS", 5),
			LoginWindowMin:   getEnvInt("LOGIN_WINDOW_MIN", 15),
			LoginBlockMin:    getEnvInt("LOGIN_BLOCK_MIN", 15),
		},
		CORS: CORSConfig{
			AllowedOrigins: origins,
		},
	}

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	return nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}

func getEnvInt64(key string, defaultValue int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	return n
}

func splitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 &&
			((strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'"))) {
			value = value[1 : len(value)-1]
		}
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
