package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Cache     CacheConfig
	RateLimit RateLimitConfig
	Token     TokenConfig
	SMTP      SMTPConfig
	Security  SecurityConfig
	CORS      CORSConfig
	Session   SessionConfig
	Tenant    TenantConfig
	TOTP      TOTPConfig
}

type AppConfig struct {
	Env      string
	Name     string
	Port     string
	LogLevel string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL             string
	Password        string
	DB              int
	MaxRetries      int
	PoolSize        int
	MinIdleConns    int
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

type CacheConfig struct {
	SessionTTL     time.Duration
	UserTTL        time.Duration
	FailedLoginTTL time.Duration
}

type RateLimitConfig struct {
	Enabled        bool
	LoginPerMin    int
	RegisterPerMin int
	APIPerMin      int
}

type TokenConfig struct {
	AccessTokenSecret    string
	RefreshTokenSecret   string
	PasetoSymmetricKey   string
	Issuer               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	MagicLinkDuration    time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type SecurityConfig struct {
	BCryptCost               int
	Argon2Memory             uint32
	Argon2Iterations         uint32
	Argon2Parallelism        uint8
	Argon2SaltLength         uint32
	Argon2KeyLength          uint32
	MaxFailedAttempts        int
	FailedLoginBlockDuration time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type SessionConfig struct {
	CookieName     string
	CookieDomain   string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite string
	MaxDevices     int
}

type TenantConfig struct {
	HeaderName      string
	DefaultTenantID string
}

type TOTPConfig struct {
	Issuer    string
	Period    uint
	Digits    int
	Algorithm string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			Name:     getEnv("APP_NAME", "auth-service"),
			Port:     getEnv("PORT", "8080"),
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL:             getEnv("REDIS_URL", "redis://localhost:6379/0"),
			Password:        getEnv("REDIS_PASSWORD", ""),
			DB:              getEnvAsInt("REDIS_DB", 0),
			MaxRetries:      getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:        getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:    getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
			ConnMaxIdleTime: getEnvAsDuration("REDIS_CONN_MAX_IDLE_TIME", 5*time.Minute),
			DialTimeout:     getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:     getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:    getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Cache: CacheConfig{
			SessionTTL:     getEnvAsDuration("CACHE_SESSION_TTL", 15*time.Minute),
			UserTTL:        getEnvAsDuration("CACHE_USER_TTL", 5*time.Minute),
			FailedLoginTTL: getEnvAsDuration("CACHE_FAILED_LOGIN_TTL", 1*time.Hour),
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnvAsBool("RATE_LIMIT_ENABLED", true),
			LoginPerMin:    getEnvAsInt("RATE_LIMIT_LOGIN_PER_MIN", 5),
			RegisterPerMin: getEnvAsInt("RATE_LIMIT_REGISTER_PER_MIN", 3),
			APIPerMin:      getEnvAsInt("RATE_LIMIT_API_PER_MIN", 60),
		},
		Token: TokenConfig{
			AccessTokenSecret:    getEnv("ACCESS_TOKEN_SECRET", ""),
			RefreshTokenSecret:   getEnv("REFRESH_TOKEN_SECRET", ""),
			PasetoSymmetricKey:   getEnv("PASETO_SYMMETRIC_KEY", ""),
			Issuer:               getEnv("TOKEN_ISSUER", "auth-service"),
			AccessTokenDuration:  getEnvAsDuration("ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("REFRESH_TOKEN_DURATION", 168*time.Hour),
			MagicLinkDuration:    getEnvAsDuration("MAGIC_LINK_DURATION", 15*time.Minute),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnvAsInt("SMTP_PORT", 1025),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@authservice.com"),
			FromName: getEnv("SMTP_FROM_NAME", "Auth Service"),
		},
		Security: SecurityConfig{
			BCryptCost:               getEnvAsInt("BCRYPT_COST", 12),
			Argon2Memory:             uint32(getEnvAsInt("ARGON2_MEMORY", 65536)),
			Argon2Iterations:         uint32(getEnvAsInt("ARGON2_ITERATIONS", 3)),
			Argon2Parallelism:        uint8(getEnvAsInt("ARGON2_PARALLELISM", 2)),
			Argon2SaltLength:         uint32(getEnvAsInt("ARGON2_SALT_LENGTH", 16)),
			Argon2KeyLength:          uint32(getEnvAsInt("ARGON2_KEY_LENGTH", 32)),
			MaxFailedAttempts:        getEnvAsInt("MAX_FAILED_LOGIN_ATTEMPTS", 5),
			FailedLoginBlockDuration: getEnvAsDuration("FAILED_LOGIN_BLOCK_DURATION", 15*time.Minute),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", ",", []string{"*"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", ",", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", ",", []string{"Accept", "Authorization", "Content-Type"}),
			ExposedHeaders:   getEnvAsSlice("CORS_EXPOSED_HEADERS", ",", []string{}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvAsInt("CORS_MAX_AGE", 300),
		},
		Session: SessionConfig{
			CookieName:     getEnv("SESSION_COOKIE_NAME", "auth_session"),
			CookieDomain:   getEnv("SESSION_COOKIE_DOMAIN", ""),
			CookieSecure:   getEnvAsBool("SESSION_COOKIE_SECURE", false),
			CookieHTTPOnly: getEnvAsBool("SESSION_COOKIE_HTTP_ONLY", true),
			CookieSameSite: getEnv("SESSION_COOKIE_SAME_SITE", "lax"),
			MaxDevices:     getEnvAsInt("SESSION_MAX_DEVICES", 5),
		},
		Tenant: TenantConfig{
			HeaderName:      getEnv("TENANT_HEADER_NAME", "X-Tenant-ID"),
			DefaultTenantID: getEnv("DEFAULT_TENANT_ID", "default"),
		},
		TOTP: TOTPConfig{
			Issuer:    getEnv("TOTP_ISSUER", "AuthService"),
			Period:    uint(getEnvAsInt("TOTP_PERIOD", 30)),
			Digits:    getEnvAsInt("TOTP_DIGITS", 6),
			Algorithm: getEnv("TOTP_ALGORITHM", "SHA1"),
		},
	}

	// Validate required fields
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}

func getEnvAsSlice(key, separator string, defaultVal []string) []string {
	if value := os.Getenv(key); value != "" {
		return split(value, separator)
	}
	return defaultVal
}

func split(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		if trimmed := trim(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	// Simple split implementation
	result := []string{}
	current := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	result = append(result, current)
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
