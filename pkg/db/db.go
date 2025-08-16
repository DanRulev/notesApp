package db

import (
	"fmt"
	"noteApp/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(cfg config.DBCfg) (*sqlx.DB, func() error, error) {
	dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v", cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password, cfg.SSLMode)
	dbConn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}

	if err := dbConn.Ping(); err != nil {
		return nil, nil, err
	}

	close := func() error {
		return dbConn.Close()
	}

	return dbConn, close, nil
}
