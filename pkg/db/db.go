package db

import (
	"context"
	"fmt"
	"noteApp/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(cfg config.DBConfig) (*sqlx.DB, func() error, error) {
	dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v",
		cfg.Conn.Host, cfg.Conn.Port, cfg.Conn.Name, cfg.Conn.User, cfg.Conn.Password, cfg.Conn.SSL)
	dbConn, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}

	ctx, canceled := context.WithTimeout(context.Background(), 5*time.Second)
	defer canceled()

	if err := dbConn.PingContext(ctx); err != nil {
		return nil, nil, err
	}

	dbConn.SetMaxOpenConns(cfg.Cfg.MaxOpenConns)
	dbConn.SetMaxIdleConns(cfg.Cfg.MaxIdleConns)
	dbConn.SetConnMaxLifetime(cfg.Cfg.ConnMaxLifeTime)
	dbConn.SetConnMaxIdleTime(cfg.Cfg.ConnMaxIdleTime)

	close := func() error {
		return dbConn.Close()
	}

	return dbConn, close, nil
}
