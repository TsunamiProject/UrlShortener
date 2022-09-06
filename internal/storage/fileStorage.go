package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

func GetFileStorage(filePath string, baseURL string) *FileStorage {
	return &FileStorage{FilePath: filePath, BaseURL: baseURL}
}

type FileStorage struct {
	FilePath string
	BaseURL  string
}

type FileStruct struct {
	CookieValue string
	URLs        JSONURL
}

func (f *FileStorage) IsOk() error {
	if len(f.FilePath) == 0 {
		return errors.New("err")
	}
	_, err := f.Write([]byte("test"), "test")
	if err != nil {
		return err
	}
	_, err = f.Read(shorten.EncodeString([]byte("test")))
	if err != nil {
		return err
	}
	_, err = f.ReadAll("test")
	if err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Batch(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}
	var batchListBefore []BatchStructBefore
	err := json.Unmarshal(b, &batchListBefore)
	if err != nil {
		return "", err
	}
	log.Println(batchListBefore)
	var batchListAfter []BatchStructAfter
	for i := range batchListBefore {
		write, err := f.Write([]byte(batchListBefore[i].OriginalURL), authCookieValue)
		if err != nil {
			return "", err
		}
		batchListAfter = append(batchListAfter, BatchStructAfter{
			CorrelationID: batchListBefore[i].CorrelationID,
			ShortURL:      write,
		})
	}
	resp, err := json.Marshal(batchListAfter)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

//return short url from original url which must be in request body, status code and error
func (f *FileStorage) Write(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}

	file, err := os.OpenFile(f.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", nil
	}
	toFile := &FileStruct{
		CookieValue: authCookieValue,
		URLs: JSONURL{
			ShortURL:    fmt.Sprintf("%s/%s", f.BaseURL, shorten.EncodeString(b)),
			OriginalURL: string(b),
		},
	}
	res, err := json.Marshal(toFile)
	if err != nil {
		return "", nil
	}

	_, err = file.Write([]byte(fmt.Sprintf("%s\n", res)))
	if err != nil {
		err = file.Close()
		if err != nil {
			return "", err
		}
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", nil
	}

	shortenURL := fmt.Sprintf("%s/%s", f.BaseURL, shorten.EncodeString(b))

	return shortenURL, nil
}

//return original url by ID as URL param, status code and error
func (f *FileStorage) Read(shortURL string) (string, error) {
	if len(shortURL) == 0 {
		return "", errors.New("request body is empty")
	}

	file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return "", nil
	}
	scanner := bufio.NewScanner(file)

	var originalURL string
	for scanner.Scan() {
		var temp FileStruct
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		if temp.URLs.ShortURL == fmt.Sprintf("%s/%s", f.BaseURL, shortURL) {
			originalURL = temp.URLs.OriginalURL
			break
		}
	}

	if err = scanner.Err(); err != nil {
		return "", nil
	}

	err = file.Close()
	if err != nil {
		return "", nil
	}

	if originalURL == "" {
		return "", fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return originalURL, nil
}

func (f *FileStorage) ReadAll(authCookieValue string) (string, error) {
	file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return "", nil
	}
	scanner := bufio.NewScanner(file)

	var resList []JSONURL
	for scanner.Scan() {
		var temp FileStruct
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		if temp.CookieValue == authCookieValue {
			resList = append(resList, temp.URLs)
		}
	}

	if err = scanner.Err(); err != nil {
		return "", nil
	}

	err = file.Close()
	if err != nil {
		return "", nil
	}

	if len(resList) == 0 {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	resp, err := json.Marshal(resList)
	if err != nil {
		return "", nil
	}

	return string(resp), nil
}

func (f *FileStorage) Delete(authCookieValue string, deleteList []string) error {
	return nil
}
