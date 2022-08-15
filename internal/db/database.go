package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Database struct {
	db *sql.DB
}

func OpenDB(DbDsn string) *Database {
	db, err := sql.Open("pgx", DbDsn)
	if err != nil {
		log.Fatal("error with accessing to DB")
	}
	return &Database{db: db}
}

func (dbObj *Database) Ping() error {
	ctx, ctxCancel := context.WithCancel(context.Background())
	err := dbObj.db.PingContext(ctx)
	if err != nil {
		ctxCancel()
		return err
	}
	ctxCancel()
	return nil
}
