package handler

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func getRefreshTokenTTL() (time.Duration, error) {
	v, err := loadViper()
	if err != nil {
		return time.Second, err
	}

	refreshTokenTTL, err := parse(v)
	if err != nil {
		return time.Second, err
	}

	return refreshTokenTTL, nil
}

func parse(v *viper.Viper) (time.Duration, error) {
	refreshTokenTTL := v.GetDuration("auth.refresh_token_ttl")
	if refreshTokenTTL == 0 {
		return 0, fmt.Errorf("refresh_token_ttl is zero or not set")
	}
	return refreshTokenTTL, nil
}

func loadViper() (*viper.Viper, error) {
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	v := viper.New()
	v.AutomaticEnv()

	v.AddConfigPath("../../configs")
	v.SetConfigName("testing")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return v, nil
}
