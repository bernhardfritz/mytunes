package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Encde struct {
	gcm cipher.AEAD
}

func NewEncde(key []byte) (*Encde, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("error creating aes block cipher", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("error setting gcm mode", err)
		return nil, err
	}

	return &Encde{
		gcm,
	}, nil
}

func (e *Encde) Encrypt(data string) (string, error) {
	plaintext := []byte(data)

	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("error generating the nonce ", err)
		return "", err
	}

	ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)

	enc := hex.EncodeToString(ciphertext)

	return enc, nil
}

func (e *Encde) Decrypt(enc string) (string, error) {
	decodedCipherText, err := hex.DecodeString(enc)
	if err != nil {
		fmt.Println("error decoding hex", err)
		return "", err
	}

	decryptedData, err := e.gcm.Open(nil, decodedCipherText[:e.gcm.NonceSize()], decodedCipherText[e.gcm.NonceSize():], nil)
	if err != nil {
		fmt.Println("error decrypting data", err)
		return "", err
	}

	return string(decryptedData), nil
}
