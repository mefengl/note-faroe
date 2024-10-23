package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func Hash(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	hash := fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key))
	return hash, nil
}

func Verify(hash string, password string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash")
	}
	if parts[0] != "" {
		return false, errors.New("invalid hash")
	}
	if parts[1] != "argon2id" {
		return false, errors.New("invalid algorithm")
	}
	if parts[2] != "v=19" {
		return false, errors.New("unsupported hash")
	}
	var m, t, p int32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, errors.New("invalid hash")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errors.New("invalid hash")
	}
	key1, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errors.New("invalid hash")
	}
	key2 := argon2.IDKey([]byte(password), salt, 2, 19456, 1, uint32(len(key1)))
	valid := subtle.ConstantTimeCompare(key1, key2)
	return valid == 1, nil
}
