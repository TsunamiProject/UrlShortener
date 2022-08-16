package db

import (
	"database/sql"
	"fmt"
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
	err := dbObj.db.Ping()
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

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	_, err := dbObj.db.Exec(urlsTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) CreateAuthTable() error {
	authIDTableQueryString := `CREATE TABLE IF NOT EXISTS AUTHURLS(
                       AUTHID varchar(255) PRIMARY KEY
					)`
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	_, err := dbObj.db.Exec(authIDTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) InsertRow(authCookieValue string, shortURL string, originalURL string) error {
	insertRowAuthQueryString := `INSERT INTO authurls VALUES ('%s') ON CONFLICT DO NOTHING`
	insertRowURLsQueryString := `INSERT INTO noauthurls VALUES ('%s', '%s', '%s')`

	_, err := dbObj.db.Exec(fmt.Sprintf(insertRowAuthQueryString, authCookieValue))
	if err != nil {
		return err
	}

	_, err = dbObj.db.Exec(fmt.Sprintf(insertRowURLsQueryString, authCookieValue, shortURL, originalURL))
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) GetRow(authCookieValue string, shortURL string) (string, error) {
	getOriginalURLQueryString := `SELECT ORIGINALURL FROM noauthurls WHERE AUTHID='%s' AND SHORTURL='%s' LIMIT 1`
	var originalURL string

	err := dbObj.db.QueryRow(fmt.Sprintf(getOriginalURLQueryString,
		shortURL, authCookieValue)).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", err
	}

	return originalURL, nil
}

func (dbObj *Database) GetAllRows(authCookieValue string) (*sql.Rows, error) {
	getOriginalURLQueryString := `SELECT SHORTURL,ORIGINALURL FROM noauthurls WHERE AUTHID='%s'`

	rows, err := dbObj.db.Query(fmt.Sprintf(getOriginalURLQueryString,
		authCookieValue))
	if err == sql.ErrNoRows {
		return nil, err
	}

	return rows, nil
}
