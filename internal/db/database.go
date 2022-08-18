package db

import (
	"database/sql"
	"log"

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
                         	AUTHID varchar(255),
                         	SHORTURL VARCHAR(255) UNIQUE,
                         	ORIGINALURL varchar(255),
                         	FOREIGN KEY (AUTHID) REFERENCES AUTHURLS(AUTHID)
                           )`
	authIDTableQueryString = `CREATE TABLE IF NOT EXISTS AUTHURLS(
                       AUTHID varchar(255) PRIMARY KEY
					)`
	uniqueShortURLIndexQueryString = `CREATE UNIQUE INDEX IF NOT EXISTS shorturl ON noauthurls (shorturl)`
	insertRowAuthQueryString       = `INSERT INTO authurls(authid) VALUES ($1) ON CONFLICT DO NOTHING`
	insertRowURLsQueryString       = `INSERT INTO noauthurls(authid, shorturl, originalurl) VALUES ($1,$2,$3) 
								ON CONFLICT (shorturl) DO UPDATE SET shorturl=$2 RETURNING authid`
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
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	_, err := dbObj.db.Exec(urlsTableQueryString)
	if err != nil {
		return err
	}
	_, err = dbObj.db.Exec(uniqueShortURLIndexQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) CreateAuthTable() error {
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	_, err := dbObj.db.Exec(authIDTableQueryString)
	if err != nil {
		return err
	}

	return nil
}

func (dbObj *Database) InsertRow(authCookieValue string, shortURL string, originalURL string) error {
	_, err := dbObj.db.Exec(insertRowAuthQueryString, authCookieValue)
	if err != nil {
		return err
	}
	var res string
	err = dbObj.db.QueryRow(insertRowURLsQueryString, authCookieValue, shortURL, originalURL).Scan(&res)
	if res != shortURL {
		return ErrDuplicateURL
	}

	return nil
}

func (dbObj *Database) GetURLRow(shortURL string) (string, error) {
	var originalURL string

	err := dbObj.db.QueryRow(getOriginalURLQueryString,
		shortURL).Scan(&originalURL)
	if err == sql.ErrNoRows {
		return "", err
	}

	return originalURL, nil
}

func (dbObj *Database) GetAllRows(authCookieValue string) (*sql.Rows, error) {
	rows, err := dbObj.db.Query(getOriginalURLsByCookie,
		authCookieValue)
	if err == sql.ErrNoRows {
		return nil, err
	}

	return rows, nil
}

func (dbObj *Database) Batch(batchList []BatchBefore, authCookieValue string) ([]BatchAfter, error) {
	_, err := dbObj.db.Exec(insertRowAuthQueryString, authCookieValue)
	if err != nil {
		//fmt.Println("Here")
		return nil, err
	}

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
		if _, err = stmt.Exec(authCookieValue, shorten.EncodeString([]byte(batchList[i].OriginalURL)),
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
