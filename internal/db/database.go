package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

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

func (dbObj *Database) Ping(ctx context.Context) error {
	err := dbObj.db.PingContext(ctx)
	if err != nil {
		return err
	}
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
	urlsTableQueryString := `CREATE TABLE IF NOT EXISTS NOAUTHURLS (
                         	AUTHID varchar(255),
                         	SHORTURL VARCHAR(255),
                         	ORIGINALURL varchar(255),
                         	FOREIGN KEY (AUTHID) REFERENCES AUTHURLS(AUTHID)
                           )`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := dbObj.db.ExecContext(ctx, urlsTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) CreateAuthTable() error {
	authIDTableQueryString := `CREATE TABLE IF NOT EXISTS AUTHURLS(
                       AUTHID varchar(255) PRIMARY KEY
					)`
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := dbObj.db.ExecContext(ctx, authIDTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) InsertRow(authCookieValue string, shortURL string, originalURL string, ctx context.Context) error {
	insertRowAuthQueryString := `INSERT INTO authurls VALUES ('%s') ON CONFLICT DO NOTHING`
	insertRowURLsQueryString := `INSERT INTO noauthurls VALUES ('%s', '%s', '%s')`

	_, err := dbObj.db.ExecContext(ctx, fmt.Sprintf(insertRowAuthQueryString, authCookieValue))
	if err != nil {
		return err
	}

	_, err = dbObj.db.ExecContext(ctx, fmt.Sprintf(insertRowURLsQueryString, authCookieValue, shortURL, originalURL))
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) GetRow(authCookieValue string, shortURL string, ctx context.Context) (string, error) {
	getOriginalURLQueryString := `SELECT ORIGINALURL FROM noauthurls WHERE AUTHID='%s' AND SHORTURL='%s' LIMIT 1`
	var originalURL string

	err := dbObj.db.QueryRowContext(ctx, fmt.Sprintf(getOriginalURLQueryString,
		shortURL, authCookieValue)).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", err
	}

	return originalURL, nil
}

func (dbObj *Database) GetAllRows(authCookieValue string, ctx context.Context) (*sql.Rows, error) {
	getOriginalURLQueryString := `SELECT SHORTURL,ORIGINALURL FROM noauthurls WHERE AUTHID='%s'`

	rows, err := dbObj.db.QueryContext(ctx, fmt.Sprintf(getOriginalURLQueryString,
		authCookieValue))
	if err == sql.ErrNoRows {
		return nil, err
	}

	return rows, nil
}
