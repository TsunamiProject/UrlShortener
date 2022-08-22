package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gofrs/uuid"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

type Database struct {
	db *sql.DB
}

type BatchBefore struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchAfter struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

const (
	urlsTableQueryString = `CREATE TABLE IF NOT EXISTS NOAUTHURLS (
    						ID varchar(255),
                         	AUTHID varchar(255),
                         	SHORTURL VARCHAR(255),
                         	ORIGINALURL varchar(255) UNIQUE
                           )`
	insertRowURLsQueryString = `INSERT INTO noauthurls(id, authid, shorturl, originalurl) VALUES ($1,$2,$3,$4) 
									  ON CONFLICT (originalurl)
									  DO UPDATE SET authid=$2 RETURNING id`
	getOriginalURLQueryString = `SELECT ORIGINALURL FROM noauthurls WHERE SHORTURL=$1 LIMIT 1`
	getOriginalURLsByCookie   = `SELECT SHORTURL,ORIGINALURL FROM noauthurls WHERE AUTHID=$1`
)

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
	_, err := dbObj.db.Exec(urlsTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) InsertRow(authCookieValue string, shortURL string, originalURL string) error {
	var res string
	uniqueID, _ := uuid.NewV1()
	err := dbObj.db.QueryRow(insertRowURLsQueryString, uniqueID.String(), authCookieValue, shortURL, originalURL).Scan(&res)
	if res != uniqueID.String() {
		return ErrDuplicateURL
	}
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) GetURLRow(shortURL string) (string, error) {
	var originalURL string

	err := dbObj.db.QueryRow(getOriginalURLQueryString,
		shortURL).Scan(&originalURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	return originalURL, nil
}

func (dbObj *Database) GetAllRows(authCookieValue string) (*sql.Rows, error) {
	rows, err := dbObj.db.Query(getOriginalURLsByCookie,
		authCookieValue)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return rows, nil
}

func (dbObj *Database) Batch(batchList []BatchBefore, authCookieValue string) ([]BatchAfter, error) {
	tx, err := dbObj.db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(insertRowURLsQueryString)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var resList []BatchAfter
	for i := range batchList {
		uniqueID, _ := uuid.NewV1()
		if _, err = stmt.Exec(uniqueID.String(), authCookieValue, shorten.EncodeString([]byte(batchList[i].OriginalURL)),
			batchList[i].OriginalURL); err != nil {
			return nil, err
		}
		resList = append(resList, BatchAfter{
			CorrelationID: batchList[i].CorrelationID,
			ShortURL:      shorten.EncodeString([]byte(batchList[i].OriginalURL)),
		})
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
		return nil, err
	}

	return resList, nil
}
