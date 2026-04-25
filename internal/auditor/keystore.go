package auditor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

type KeyStore struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func NewKeyStore(keyPath string, passphrase string) (*KeyStore, error) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return generateAndSave(keyPath, passphrase)
	}
	return loadAndDecrypt(keyPath, passphrase)
}

func generateAndSave(path string, passphrase string) (*KeyStore, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := deriveKey(passphrase, salt)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nil, nonce, priv, nil)
	
	// Format: [salt(16)][nonce(GCM)][ciphertext]
	final := append(salt, append(nonce, encrypted...)...)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, final, 0600); err != nil {
		return nil, err
	}

	return &KeyStore{pub: pub, priv: priv}, nil
}

func loadAndDecrypt(path string, passphrase string) (*KeyStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) < 16+12 { // salt + min nonce
		return nil, errors.New("malformed key file")
	}

	salt := data[:16]
	key := deriveKey(passphrase, salt)
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce := data[16 : 16+nonceSize]
	ciphertext := data[16+nonceSize:]

	privRaw, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt key - incorrect passphrase?")
	}

	priv := ed25519.PrivateKey(privRaw)
	return &KeyStore{
		priv: priv,
		pub:  priv.Public().(ed25519.PublicKey),
	}, nil
}

func deriveKey(passphrase string, salt []byte) []byte {
	// Argon2id parameters (Production-grade)
	return argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)
}

func (ks *KeyStore) Sign(message []byte) []byte {
	return ed25519.Sign(ks.priv, message)
}

func (ks *KeyStore) PublicKey() ed25519.PublicKey {
	return ks.pub
}
