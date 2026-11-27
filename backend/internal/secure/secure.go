package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
)

// prefixVersion menandai ciphertext hasil enkripsi. Nilai tanpa prefix dianggap
// data lama (plaintext) agar session yang sudah ada tetap terbaca.
const prefixVersion = "v1:"

// minSecretLength adalah panjang minimum SESSION_SECRET untuk mode produksi.
const minSecretLength = 32

var aead cipher.AEAD

// Init menginisialisasi kunci enkripsi dari SESSION_SECRET.
// Kunci diturunkan dengan SHA-256 sehingga panjang input apa pun aman.
func Init(secret string) error {
	if strings.TrimSpace(secret) == "" {
		log.Printf("WARNING: SESSION_SECRET is empty; using an insecure default key that must not be used in production")
		secret = "only4.fun-dev-insecure-default-key"
	}
	sum := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	aead = gcm
	return nil
}

// LoadFromEnv adalah jalan pintas untuk Init(os.Getenv("SESSION_SECRET")).
// Dipakai test dan mode dev; production memakai LoadFromEnvWithPolicy.
func LoadFromEnv() error {
	return Init(os.Getenv("SESSION_SECRET"))
}

// LoadFromEnvWithPolicy menolak secret kosong atau pendek kecuali allowInsecure
// (DEV_LOGIN) aktif. Ini mencegah production diam-diam memakai kunci fallback
// yang sudah diketahui publik.
func LoadFromEnvWithPolicy(allowInsecure bool) error {
	secret := os.Getenv("SESSION_SECRET")
	if !allowInsecure {
		if err := ValidateSecret(secret); err != nil {
			return err
		}
	}
	return Init(secret)
}

// ValidateSecret memeriksa panjang SESSION_SECRET untuk production.
func ValidateSecret(secret string) error {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return fmt.Errorf("SESSION_SECRET is required in production (generate with: openssl rand -hex 32)")
	}
	if len(trimmed) < minSecretLength {
		return fmt.Errorf("SESSION_SECRET must be at least %d characters in production", minSecretLength)
	}
	return nil
}

// Encrypt mengenkripsi plaintext dengan AES-256-GCM.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if aead == nil {
		return "", fmt.Errorf("encryption key is not initialized")
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefixVersion + base64.RawStdEncoding.EncodeToString(sealed), nil
}

// Decrypt membuka ciphertext. Nilai lama tanpa prefix dikembalikan apa adanya.
func Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, prefixVersion) {
		return value, nil
	}
	if aead == nil {
		return "", fmt.Errorf("encryption key is not initialized")
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, prefixVersion))
	if err != nil {
		return "", fmt.Errorf("corrupted ciphertext: %w", err)
	}
	if len(raw) < aead.NonceSize() {
		return "", fmt.Errorf("ciphertext is too short")
	}
	nonce, data := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plain), nil
}

// HashID menghasilkan hash satu arah untuk ID session/token acak.
// Nilai asli tidak pernah disimpan di database.
func HashID(id string) string {
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}

// NewSecret membuat secret acak 256-bit (hex) untuk webhook/keperluan lain.
func NewSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// NewSessionID membuat ID session acak 256-bit.
func NewSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sess_" + hex.EncodeToString(buf), nil
}
