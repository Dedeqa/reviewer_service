package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(dsn string) error {
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return DB.PingContext(ctx)
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
