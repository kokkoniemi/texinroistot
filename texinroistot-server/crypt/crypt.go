package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kokkoniemi/texinroistot/config"
)

func Encrypt(input string) (Encrypted, error) {
	var encrypted Encrypted
	iv, err := RandomBytes(aes.BlockSize)

	if err != nil {
		return encrypted, fmt.Errorf("[Encrypt] creating nonce failed: %w", err)
	}

	secret := []byte(config.Secret)
	block, err := aes.NewCipher(secret)
	if err != nil {
		return encrypted, fmt.Errorf("[Encrypt] creating cipher failed: %w", err)
	}

	plainText := []byte(input)
	cipherText := make([]byte, len(plainText))

	ctr := cipher.NewCTR(block, iv)
	ctr.XORKeyStream(cipherText, plainText)
	encrypted.iv = hex.EncodeToString(iv)
	encrypted.content = hex.EncodeToString(cipherText)

	return encrypted, nil
}

func Decrypt(encrypted *Encrypted) (string, error) {
	secret := []byte(config.Secret)

	cipherText, err := hex.DecodeString(encrypted.content)
	if err != nil {
		return "", fmt.Errorf("[Decrypt] decoding content to []byte failed: %w", err)
	}

	iv, err := hex.DecodeString(encrypted.iv)
	if err != nil {
		return "", fmt.Errorf("[Decrypt] decoding iv to []byte failed: %w", err)
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", fmt.Errorf("[Decrypt] creating cipher failed: %w", err)
	}

	plainText := make([]byte, len(cipherText))

	ctr := cipher.NewCTR(block, iv)
	ctr.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}

func Hash(input string) string {
	hash := hashSha256(input)
	iterations := 10

	for {
		hash = hashSha256(hash)

		iterations--
		if iterations < 1 {
			return hash
		}
	}
}

func hashSha256(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input + config.Salt))
	return hex.EncodeToString(hasher.Sum(nil))
}

func RandomBytes(num int) ([]byte, error) {
	if num <= 0 || num > 9999 {
		return nil, fmt.Errorf("[RandomBytes] invalid argument 'num': %d. It must be > 0 and < 9999", num)
	}
	bytes := make([]byte, num)
	_, err := rand.Read(bytes)

	if err != nil {
		return nil, fmt.Errorf("[RandomBytes] internal error: %w", err)
	}

	return bytes, nil
}
