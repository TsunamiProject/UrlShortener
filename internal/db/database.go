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
    						TODELETE BOOLEAN,
                         	AUTHID varchar(255),
                         	SHORTURL VARCHAR(255),
                         	ORIGINALURL varchar(255) UNIQUE
                           )`
	insertRowURLsQueryString = `INSERT INTO noauthurls(id, todelete, authid, shorturl, originalurl) 
									  VALUES ($1,$2,$3,$4,$5) 
									  ON CONFLICT (originalurl)
									  DO UPDATE SET authid=$3 RETURNING id`
	deleteRowByUser           = `UPDATE noauthurls set todelete=$1 where authid=$2 and shorturl=$3 returning todelete`
	getOriginalURLQueryString = `SELECT ORIGINALURL,TODELETE FROM noauthurls WHERE SHORTURL=$1 LIMIT 1`
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
	err := dbObj.db.QueryRow(insertRowURLsQueryString, uniqueID.String(), false, authCookieValue,
		shortURL, originalURL).Scan(&res)
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
	var toDelete bool

	err := dbObj.db.QueryRow(getOriginalURLQueryString,
		shortURL).Scan(&originalURL, &toDelete)
	if errors.Is(err, sql.ErrNoRows) {
		log.Println(originalURL)
		log.Println(toDelete)
		return "", err
	}
	if toDelete == true {
		return "", ErrDeletedURL
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
		if _, err = stmt.Exec(uniqueID.String(), false, authCookieValue, shorten.EncodeString([]byte(batchList[i].OriginalURL)),
			batchList[i].OriginalURL); err != nil {
			return nil, err
		}
		resList = append(resList, BatchAfter{
			CorrelationID: batchList[i].CorrelationID,
			ShortURL:      shorten.EncodeString([]byte(batchList[i].OriginalURL)),
		})
	}

	if err := tx.Commit(); err != nil {
		log.Printf("update drivers: unable to commit: %v\n", err)
		return nil, err
	}

	return resList, nil
}

func (dbObj *Database) DeleteRows(authCookieValue string, shortURLS []string) error {
	tx, err := dbObj.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(deleteRowByUser)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for shortURL := range shortURLS {
		log.Println("URL:", shortURLS[shortURL])
		_, err = stmt.Exec(true, authCookieValue, shortURLS[shortURL])
		if err != nil {
			log.Println("ERR is:", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("update drivers: unable to commit: %v\n", err)
		return err
	}

	return nil
}
