package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Vault struct{ aead cipher.AEAD }

// Open keeps the encryption key separate from DB backups and binds every
// ciphertext to the secret's immutable identifier using authenticated data.
func Open(dataDir, encodedKey string) (*Vault, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	keyFile := filepath.Join(dataDir, "encryption.key")
	if encodedKey == "" {
		b, err := os.ReadFile(keyFile)
		if errors.Is(err, os.ErrNotExist) {
			key := make([]byte, 32)
			if _, err = rand.Read(key); err != nil {
				return nil, err
			}
			encodedKey = base64.StdEncoding.EncodeToString(key)
			f, e := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if e != nil {
				return nil, e
			}
			_, e = f.WriteString(encodedKey + "\n")
			if e == nil {
				e = f.Sync()
			}
			closeErr := f.Close()
			if e != nil {
				return nil, e
			}
			if closeErr != nil {
				return nil, closeErr
			}
		} else if err != nil {
			return nil, err
		} else {
			encodedKey = strings.TrimSpace(string(b))
		}
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 base64-encoded bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, err
	}
	return &Vault{aead: aead}, nil
}
func (v *Vault) Encrypt(id, value string) string {
	return base64.StdEncoding.EncodeToString(v.aead.Seal(
		nil,
		nil,
		[]byte(value),
		[]byte(id),
	))
}
func (v *Vault) Decrypt(id, value string) (string, error) {
	b, e := base64.StdEncoding.DecodeString(value)
	if e != nil {
		return "", errors.New("invalid encrypted secret")
	}
	p, e := v.aead.Open(
		nil,
		nil,
		b,
		[]byte(id),
	)
	if e != nil {
		return "", errors.New("secret authentication failed")
	}
	return string(p), nil
}
func Token() string {
	var b [32]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
func Equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func HashPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > 72 {
		return "", errors.New("password must contain 12 to 72 bytes")
	}
	b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), e
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
