package config

import (
	"fmt"
	"noteApp/pkg/valid"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type LoggerCfg struct {
	Level             string   `mapstructure:"level"`
	Development       bool     `mapstructure:"development"`
	DisableCaller     bool     `mapstructure:"disable_caller"`
	DisableStacktrace bool     `mapstructure:"disable_stacktrace"`
	Encoding          string   `mapstructure:"encoding"`
	OutputPaths       []string `mapstructure:"output_paths"`
	ErrorOutputPaths  []string `mapstructure:"error_output_paths"`
}

type DBCfg struct {
	Driver   string `mapstructure:"driver" validate:"required"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	SSLMode  string `mapstructure:"sslmode" validate:"required"`
	DSN      string `mapstructure:"dsn"`
}

type ServerCfg struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            string        `mapstructure:"port" validate:"required"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" validate:"required"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes" validate:"required"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" validate:"required"`
}

type AuthCfg struct {
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" validate:"required"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" validate:"required"`
	JwtSecret       string        `mapstructure:"jwt_secret" validate:"required"`
}

type Config struct {
	Auth   AuthCfg   `mapstructure:"auth"`
	DB     DBCfg     `mapstructure:"db"`
	Server ServerCfg `mapstructure:"server"`
	Logger LoggerCfg `mapstructure:"logger"`
}

func InitConfig() (*Config, error) {
	v, err := loadViper()
	if err != nil {
		return nil, fmt.Errorf("failed load viper: %w", err)
	}

	cfg, err := parseConfig(v)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := valid.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	return cfg, err
}

func loadViper() (*viper.Viper, error) {
	v := viper.New()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	v.AutomaticEnv()

	configName := os.Getenv("CONFIG_NAME")
	if configName == "" {
		configName = "testing"
	}

	v.AddConfigPath("configs")
	v.SetConfigName(configName)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return v, nil
}

func parseConfig(v *viper.Viper) (*Config, error) {
	cfg := Config{}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
