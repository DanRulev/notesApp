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

	driver := os.Getenv("DB_DRIVER")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if driver == "" {
		return "", "", fmt.Errorf("driver missing")
	}
	if port == "" {
		return "", "", fmt.Errorf("port missing")
	}
	if name == "" {
		return "", "", fmt.Errorf("name missing")
	}
	if user == "" {
		return "", "", fmt.Errorf("user missing")
	}
	if password == "" {
		return "", "", fmt.Errorf("password missing")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, name)
	return driver, dsn, nil
}
