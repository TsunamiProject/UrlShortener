package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	cipherPass = "practicum"
)

type Cookier struct {
	Encrypted []byte
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Cookier) encodeCookie() error {
	key := sha256.Sum256([]byte(cipherPass))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	src, err := generateRandom(aesgcm.NonceSize())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	dst := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем
	c.Encrypted = dst

	return nil
}

func (c *Cookier) decodeCookie() error {
	key := sha256.Sum256([]byte(cipherPass))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	_, err = aesgcm.Open(nil, nonce, c.Encrypted, nil)
	if err != nil {
		return err
	}

	return nil
}

func CreateNewCookie(c *Cookier) (*http.Cookie, error) {
	err := c.encodeCookie()
	if err != nil {
		return nil, err
	}
	newCookie := http.Cookie{
		Name:    "auth",
		Value:   hex.EncodeToString(c.Encrypted),
		Expires: time.Time{},
	}
	return &newCookie, nil
}

func CookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		ck := &Cookier{}
		cookieObj := &http.Cookie{}
		if err != nil {
			newCookie, err := CreateNewCookie(ck)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cookieObj = newCookie
			//http.SetCookie(w, cookie)
			//r.AddCookie(cookie)
		} else {
			decodedStr, err := hex.DecodeString(cookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ck.Encrypted = decodedStr
			err = ck.decodeCookie()
			if err != nil {
				newCookie, err := CreateNewCookie(ck)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				cookieObj = newCookie
				//http.SetCookie(w, newCookie)
				//r.AddCookie(newCookie)
			}
			//newCookie := http.Cookie{
			//	Name:    "auth",
			//	Value:   hex.EncodeToString(ck.Encrypted),
			//	Expires: time.Time{},
			//}
			//http.SetCookie(w, &newCookie)
			//r.AddCookie(&newCookie)
		}
		http.SetCookie(w, cookieObj)
		r.AddCookie(cookieObj)

		next.ServeHTTP(w, r)
	})
}
