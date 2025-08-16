package repository

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var globalTestDB *sqlx.DB

func TestMain(m *testing.M) {
	var err error
	globalTestDB, err = initGlobalTestDB()
	if err != nil {
		fmt.Printf("failed to init db: %v\n", err)
		os.Exit(1)
	}

	exit := m.Run()

	if err := globalTestDB.Close(); err != nil {
		fmt.Printf("failed to close db: %v\n", err)
		os.Exit(1)
	}

	os.Exit(exit)
}

func initGlobalTestDB() (*sqlx.DB, error) {
	driverName, dsn, err := getConnectParam()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getConnectParam() (string, string, error) {
	if err := godotenv.Load("../../.env"); err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to load .env file: %w", err)
	}

	v := viper.New()
	v.AutomaticEnv()

	driver := viper.GetString("DB_DRIVER")
	host := "localhost"
	port := viper.GetString("DB_PORT")
	name := viper.GetString("DB_NAME")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")

	if driver == "" {
		return "", "", fmt.Errorf("driver env var missing")
	}

	if port == "" {
		return "", "", fmt.Errorf("port env var missing")
	}

	if name == "" {
		return "", "", fmt.Errorf("name env var missing")
	}

	if user == "" {
		return "", "", fmt.Errorf("user env var missing")
	}

	if password == "" {
		return "", "", fmt.Errorf("password env var missing")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)

	return driver, dsn, nil
}
