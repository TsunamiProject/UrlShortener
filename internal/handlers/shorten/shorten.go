package shorten

import (
	"crypto/sha1"
	"encoding/hex"
)

func HashString(b []byte) string {
	sha1Obj := sha1.New()
	sha1Obj.Write(b)
	sha1Hash := hex.EncodeToString(sha1Obj.Sum(nil))
	return sha1Hash[:len(sha1Hash)/2]
}
