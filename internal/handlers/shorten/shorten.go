package shorten

import (
	"encoding/hex"
)

func EncodeString(b []byte) string {
	encodedStr := hex.EncodeToString(b)
	return encodedStr
}

func DecodeString(b []byte) []byte {
	decodedString, _ := hex.DecodeString(string(b))
	return decodedString
}
