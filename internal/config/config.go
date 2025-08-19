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

type DBConfig struct {
	Conn DBConn `mapstructure:"conn"`
	Cfg  DBCfg  `mapstructure:"cfg"`
}

type DBConn struct {
	Driver   string `mapstructure:"driver" validate:"required"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
	SSL      string `mapstructure:"ssl" validate:"oneof=disable require verify-full"`
}

type DBCfg struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1,max=1000"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=0,max=100"`
	ConnMaxLifeTime time.Duration `mapstructure:"conn_max_life_time" validate:"min=0"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"min=0"`
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
	DB     DBConfig  `mapstructure:"db"`
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

	err := v.BindEnv("db.conn.driver", "DB_CONN_DRIVER")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_DRIVER: %w", err)
	}
	err = v.BindEnv("db.conn.host", "DB_CONN_HOST")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_HOST: %w", err)
	}
	err = v.BindEnv("db.conn.port", "DB_CONN_PORT")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_PORT: %w", err)
	}
	err = v.BindEnv("db.conn.user", "DB_CONN_USER")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_USER: %w", err)
	}
	err = v.BindEnv("db.conn.password", "DB_CONN_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_PASSWORD: %w", err)
	}
	err = v.BindEnv("db.conn.name", "DB_CONN_NAME")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_NAME: %w", err)
	}
	err = v.BindEnv("db.conn.ssl", "DB_CONN_SSL")
	if err != nil {
		return nil, fmt.Errorf("fail bind DB_CONN_SSL: %w", err)
	}

	err = v.BindEnv("auth.jwt_secret", "AUTH_JWT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("fail bind AUTH_JWT_SECRET: %w", err)
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
