// Package crypt provides the local device-key cryptography used by the client.
package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

const (
	defaultDeviceKeyKDFAlgorithm  = "pbkdf2-sha256"
	defaultDeviceKeyKDFIterations = 600000
	defaultDeviceKeyLength        = 32
	deviceKeySaltLength           = 16
	deviceKeyNonceLength          = 12
)

type DeviceKeyKDFParams struct {
	Algorithm  string `json:"algorithm"`
	Iterations int    `json:"iterations"`
	KeyLength  int    `json:"key_length"`
}

type EncryptedDeviceKeyRecord struct {
	EncryptedPrivateKey []byte
	KDFSalt             []byte
	KDFParams           string
}

func DefaultDeviceKeyKDFParams() DeviceKeyKDFParams {
	return DeviceKeyKDFParams{
		Algorithm:  defaultDeviceKeyKDFAlgorithm,
		Iterations: defaultDeviceKeyKDFIterations,
		KeyLength:  defaultDeviceKeyLength,
	}
}

func GenerateDeviceKeyPair() ([]byte, []byte, error) {
	privateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return privateKey.Bytes(), privateKey.PublicKey().Bytes(), nil
}

func GenerateDeviceKeySalt() ([]byte, error) {
	return randomBytes(deviceKeySaltLength)
}

func MarshalKDFParams(params DeviceKeyKDFParams) (string, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("marshal KDF params: %w", err)
	}

	return string(data), nil
}

func ParseKDFParams(raw string) (DeviceKeyKDFParams, error) {
	var params DeviceKeyKDFParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return DeviceKeyKDFParams{}, fmt.Errorf("parse KDF params: %w", err)
	}
	if params.Algorithm != defaultDeviceKeyKDFAlgorithm {
		return DeviceKeyKDFParams{}, fmt.Errorf("unsupported KDF algorithm %q", params.Algorithm)
	}
	if params.Iterations <= 0 || params.KeyLength <= 0 {
		return DeviceKeyKDFParams{}, fmt.Errorf("invalid KDF params")
	}

	return params, nil
}

func EncryptDevicePrivateKey(privateKey []byte, password string, salt []byte, params DeviceKeyKDFParams) ([]byte, error) {
	derivedKey, err := deriveDeviceEncryptionKey(password, salt, params)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := randomBytes(deviceKeyNonceLength)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, privateKey, nil)
	return append(nonce, ciphertext...), nil
}

func DecryptDevicePrivateKey(record EncryptedDeviceKeyRecord, password string) ([]byte, error) {
	params, err := ParseKDFParams(record.KDFParams)
	if err != nil {
		return nil, err
	}

	if len(record.EncryptedPrivateKey) <= deviceKeyNonceLength {
		return nil, fmt.Errorf("encrypted device private key is invalid")
	}

	derivedKey, err := deriveDeviceEncryptionKey(password, record.KDFSalt, params)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := record.EncryptedPrivateKey[:deviceKeyNonceLength]
	ciphertext := record.EncryptedPrivateKey[deviceKeyNonceLength:]
	privateKey, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func deriveDeviceEncryptionKey(password string, salt []byte, params DeviceKeyKDFParams) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, params.Iterations, params.KeyLength)
}

func randomBytes(length int) ([]byte, error) {
	buffer := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
