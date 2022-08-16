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

func ConnectToDB(databaseDsn string) *Database {
	db, err := sql.Open("pgx", databaseDsn)
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

func (dbObj *Database) CloseDBConn() error {
	err := dbObj.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (dbObj *Database) CreateURLsTable() error {
	urlsTable := `CREATE TABLE IF NOT EXISTS NOAUTHURLS (
                         	AUTHID varchar(255),
                         	SHORTURL VARCHAR(255),
                         	ORIGINALURL varchar(255),
                         	FOREIGN KEY (AUTHID) REFERENCES AUTHURLS(AUTHID)
                           )`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := dbObj.db.ExecContext(ctx, urlsTable)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) CreateAuthTable() error {
	authIDTable := `CREATE TABLE IF NOT EXISTS AUTHURLS(
                       AUTHID varchar(255) PRIMARY KEY
					)`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := dbObj.db.ExecContext(ctx, authIDTable)
	if err != nil {
		return err
	}

	return nil
}
