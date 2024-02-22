// Package encryptor implements RSA encryption/decryption.
//
// Intention of the package is to be used in ecnrypted communication between
// Agent and Server modules. Messages from the Agent, before being sent, are
// going to be encrypted (by public key) with Encryptor and then decrypted (by
// private key) with Decryptor on recipient side - the Server.
//
// A pair of keys must be generated on the Server's side:
//   - public key is provided to the Agent for messages to be encrypted with;
//   - private key is used by the Server to decrypt data sent from the Agent.
//
// TODO: Signing (sign, then encrypt) should also be implemented in such form of
// communication.
//
// Alternative encryption packages:
//   - https://github.com/golang-module/dongle
//   - https://github.com/avelino/awesome-go#security
package encryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// Supported pem block type.
const (
	BlockTypePKCS1PrivateKey = "RSA PRIVATE KEY"
	BlockTypePKCS1PublicKey  = "RSA PUBLIC KEY"
	BlockTypePKCS8PrivateKey = "PRIVATE KEY"
	BlockTypePKCS8PublicKey  = "PUBLIC KEY"
)

// Encryptor implements encryption.
type Encryptor struct {
	publicKey *rsa.PublicKey
}

// Decryptor implements decryption of an encrypted messages.
type Decryptor struct {
	privateKey *rsa.PrivateKey
}

// NewEncryptor creates new Encryptor instance.
// Requires path to file with RSA public key.
func NewEncryptor(public string) (*Encryptor, error) {
	// read from file
	fPub, err := os.ReadFile(public)
	if err != nil {
		return nil, fmt.Errorf("failed to read public encryption key from file '%s', err: %v", public, err)
	}

	// parse public key from bytes
	pub, err := parsePublicKey(fPub)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v, file '%s'", err, public)
	}

	return &Encryptor{
		publicKey: pub,
	}, nil
}

// NewDecryptor creates new Decryptor instance.
// Requires path to file with RSA private key.
func NewDecryptor(private string) (*Decryptor, error) {
	// read from file
	fPriv, err := os.ReadFile(private)
	if err != nil {
		return nil, fmt.Errorf("failed to read private encryption key from file '%s': %v", private, err)
	}

	// parse private key from bytes
	priv, err := parsePrivateKey(fPriv)
	if err != nil {
		return nil, fmt.Errorf("parse error: %v, file '%s'", err, private)
	}

	return &Decryptor{
		privateKey: priv,
	}, nil
}

// Encrypt encrypts message.
func (e *Encryptor) Encrypt(msg []byte) (data []byte, err error) {
	// Another way to implement batching (splitting in chunks)
	// but that approach allocates [][]byte slice:
	// https://go.dev/wiki/SliceTricks#batching-with-minimal-allocation

	// encryption/decryption is implemented in chunks to avoid
	// "crypto/rsa: message too long for RSA key size" error

	if len(msg) == 0 {
		return data, nil
	}

	hash := sha256.New()

	// chunk step
	step := e.publicKey.Size() - 2*hash.Size() - 2

	// encrypted part of data
	var chunk []byte

	for i, end := 0, 0; i < len(msg); {
		end = i + step
		if end > len(msg) {
			end = len(msg)
		}

		chunk, err = rsa.EncryptOAEP(hash, rand.Reader, e.publicKey, msg[i:end], nil)
		if err != nil {
			return nil, err
		}

		data = append(data, chunk...)
		i = end
	}

	return data, nil
}

// Decrypt decrypts ciphertext data.
func (e *Decryptor) Decrypt(data []byte) (msg []byte, err error) {
	// Another way to implement batching (splitting in chunks)
	// but that approach allocates [][]chunk slice:
	// https://go.dev/wiki/SliceTricks#batching-with-minimal-allocation

	// encryption/decryption is implemented in chunks to avoid
	// "crypto/rsa: message too long for RSA key size" error

	if len(data) == 0 {
		return msg, nil
	}

	hash := sha256.New()

	// chunk step
	step := e.privateKey.Size()

	// decrypted part of data
	var chunk []byte

	for i := 0; i < len(data); {
		end := i + step
		if end > len(data) {
			end = len(data)
		}
		chunk, err = rsa.DecryptOAEP(hash, rand.Reader, e.privateKey, data[i:end], nil)
		if err != nil {
			return nil, err
		}
		msg = append(msg, chunk...)
		i = end
	}

	return msg, nil
}

// parsePublicKey parses public key encoded in PEM form.
// PKCS1 and PKIX (PKCS8) are supported.
func parsePublicKey(public []byte) (*rsa.PublicKey, error) {
	var (
		block  *pem.Block
		keyAny any
		err    error
	)

	if block, _ = pem.Decode(public); block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	switch block.Type {
	case BlockTypePKCS1PublicKey:
		keyAny, err = x509.ParsePKCS1PublicKey(block.Bytes)
	// XXX: BlockTypePKCS8PublicKey check might be not needed
	// and better used as default case?
	// Because, as I understand, x509.ParsePKIXPublicKey covers everything.
	case BlockTypePKCS8PublicKey:
		keyAny, err = x509.ParsePKIXPublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unexpected pem block type: '%s'", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key, err: %v", err)
	}

	publicKey, ok := keyAny.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsing ended up with unexpected key type while wanted *rsa.PublicKey")
	}

	return publicKey, nil
}

// parsePrivateKey parses private key encoded in PEM form.
// PKCS1 and PKCS8 are supported only.
func parsePrivateKey(private []byte) (*rsa.PrivateKey, error) {
	var (
		block  *pem.Block
		keyAny any
		err    error
	)

	if block, _ = pem.Decode(private); block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	switch block.Type {
	case BlockTypePKCS1PrivateKey:
		keyAny, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case BlockTypePKCS8PrivateKey:
		keyAny, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unexpected pem block type: '%s'", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key, err: %v", err)
	}

	privateKey, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parsing ended up with unexpected key type while wanted *rsa.PrivateKey")
	}

	return privateKey, nil
}
