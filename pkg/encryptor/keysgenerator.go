package encryptor

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GenerateKeys generates a random RSA private key of the given bit size.
func GenerateKeys(size int) (private *rsa.PrivateKey, err error) {
	return rsa.GenerateKey(rand.Reader, size)
}

// FormatPrivateKey encodes private key to PEM format.
func FormatPrivateKey(private *rsa.PrivateKey) (keyPEM []byte, err error) {

	if keyPEM, err = x509.MarshalPKCS8PrivateKey(private); err != nil {
		return nil, err
	}

	var keyBuf bytes.Buffer
	if err = pem.Encode(&keyBuf, &pem.Block{
		Type:  BlockTypePKCS8PrivateKey,
		Bytes: keyPEM,
	}); err != nil {
		return nil, err
	}

	return keyBuf.Bytes(), err
}

// FormatPublicKey encodes public key to PEM format.
func FormatPublicKey(public *rsa.PublicKey) (keyPEM []byte, err error) {

	if keyPEM, err = x509.MarshalPKIXPublicKey(public); err != nil {
		return nil, err
	}

	var keyBuf bytes.Buffer
	if err = pem.Encode(&keyBuf, &pem.Block{
		Type:  BlockTypePKCS8PublicKey,
		Bytes: keyPEM,
	}); err != nil {
		return nil, err
	}

	return keyBuf.Bytes(), err
}
