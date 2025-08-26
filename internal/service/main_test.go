package service

import (
	"fmt"
	"noteApp/internal/config"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func initConfig() (config.AuthCfg, error) {
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		return config.AuthCfg{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	v := viper.New()
	v.AutomaticEnv()

	jwtSecret := v.GetString("JWT_SECRET")

	if jwtSecret == "" {
		return config.AuthCfg{}, fmt.Errorf("jwt secret is not set")
	}

	v.AddConfigPath("../../configs")
	v.SetConfigName("testing")

	if err := v.ReadInConfig(); err != nil {
		return config.AuthCfg{}, fmt.Errorf("failed read in config: %w", err)
	}

	var cfg config.AuthCfg
	if err := v.UnmarshalKey("auth", &cfg); err != nil {
		return config.AuthCfg{}, err
	}

	cfg.JwtSecret = jwtSecret

	return cfg, nil
}
