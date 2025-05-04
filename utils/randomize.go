package utils

import (
	cr "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.NewSource(time.Now().UnixNano())
}
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(stringSize int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < stringSize; i++ {
		character := alphabet[rand.Intn(k)]
		sb.WriteByte(character)
	}
	return sb.String()
}
func RandomUsername() string {
	return RandomString(5)
}
func RandomEmail() string {
	return fmt.Sprintf("%v@gmail.com", RandomUsername())
}
func RandomByte(size int) ([]byte, error) {
	byte := make([]byte, size)
	_, err := cr.Read(byte)
	if err != nil {
		return nil, err
	}
	return byte, nil
}
func HashRandomBytes(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}
