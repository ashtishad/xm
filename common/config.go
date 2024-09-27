package common

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/ashtishad/xm/internal/security"
	"github.com/spf13/viper"
)

// AppConfig centralizes all configuration settings for the application.
type AppConfig struct {
	DB     DBConfig
	JWT    *security.JWTManager
	Server ServerConfig
}

// DBConfig holds database connection parameters.
type DBConfig struct {
	ConnString      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ServerConfig contains server-specific settings.
type ServerConfig struct {
	Address string
	GinMode string
}

// LoadConfig initializes the application configuration.
func LoadConfig(l *slog.Logger) (*AppConfig, error) {
	v := viper.New()
	setupViper(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := checkRequiredConfigs(v); err != nil {
		return nil, err
	}

	config := &AppConfig{
		DB: DBConfig{
			ConnString:      v.GetString("DB_CONN_STRING"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
			ConnMaxIdleTime: v.GetDuration("DB_CONN_MAX_IDLE_TIME"),
		},
		Server: ServerConfig{
			Address: v.GetString("SERVER_ADDRESS"),
			GinMode: v.GetString("GIN_MODE"),
		},
	}

	var err error
	if config.JWT, err = setupJWTManager(v); err != nil {
		return nil, fmt.Errorf("failed to setup JWT manager: %w", err)
	}

	logConfig(l, config)
	return config, nil
}

// setupViper configures Viper to read from the env file.
// Allows environment variables to override config file
func setupViper(v *viper.Viper) {
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()
}

// checkRequiredConfigs ensures that all required configuration variables are present.
func checkRequiredConfigs(v *viper.Viper) error {
	requiredConfigs := []string{
		"DB_CONN_STRING", "JWT_PRIVATE_KEY", "JWT_PUBLIC_KEY",
		"SERVER_ADDRESS", "GIN_MODE", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME", "DB_CONN_MAX_IDLE_TIME",
	}
	var missingConfigs []string

	for _, rc := range requiredConfigs {
		if !v.IsSet(rc) {
			missingConfigs = append(missingConfigs, rc)
		}
	}

	if len(missingConfigs) > 0 {
		return fmt.Errorf("missing required configuration variables: %v", missingConfigs)
	}

	return nil
}

// setupJWTManager creates a new JWTManager with decoded keys from the configuration.
// default access token duration 30 mins
func setupJWTManager(v *viper.Viper) (*security.JWTManager, error) {
	privateKeyB64 := v.GetString("JWT_PRIVATE_KEY")
	publicKeyB64 := v.GetString("JWT_PUBLIC_KEY")

	privateKey, err := decodeBase64(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT private key: %w", err)
	}

	publicKey, err := decodeBase64(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT public key: %w", err)
	}

	return security.NewJWTManager(30*time.Minute, privateKey, publicKey)
}

// decodeBase64 decodes a base64-encoded string.
func decodeBase64(encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}
	return decoded, nil
}

// logConfig logs the non-sensitive parts of the configuration.
func logConfig(l *slog.Logger, config *AppConfig) {
	l.Info("Application configuration loaded",
		"server_address", config.Server.Address,
		"gin_mode", config.Server.GinMode,
		"db_max_open_conns", config.DB.MaxOpenConns,
		"db_max_idle_conns", config.DB.MaxIdleConns,
		"db_conn_max_lifetime", config.DB.ConnMaxLifetime,
		"db_conn_max_idle_time", config.DB.ConnMaxIdleTime,
	)
}
