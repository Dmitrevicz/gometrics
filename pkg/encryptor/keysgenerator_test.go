package encryptor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGenerateKeys tests functions that encode keys to PEM format.
func TestGenerateKeys(t *testing.T) {
	// generate rsa private key
	privateKey, err := GenerateKeys(2048)
	require.NoError(t, err, "failed to generate private key: %v", err)

	// encode keys to PEM format
	privatePEM, err := FormatPrivateKey(privateKey)
	require.NoError(t, err, "failed to encode private key to PEM format: %v", err)
	publicPEM, err := FormatPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "failed to encode private key to PEM format: %v", err)

	// write PEM-encoded keys to temporary files
	pub, priv := prepareTestGenerateKeysFiles(t, privatePEM, publicPEM)

	// create encryptor
	encryptor, err := NewEncryptor(pub)
	require.NoError(t, err, "failed to create Encryptor")

	// create decryptor
	decryptor, err := NewDecryptor(priv)
	require.NoError(t, err, "failed to create Decryptor")

	// msg to be encrypted
	msg := "Привет, World!"

	// encrypted message
	var cipher []byte

	t.Run("Encrypt", func(t *testing.T) {
		var err error
		cipher, err = encryptor.Encrypt([]byte(msg))
		require.NoError(t, err)
		require.NotEmpty(t, cipher, "Encryption attempt returned no data")
		require.NotContains(t, string(cipher), msg, "Encryptor didn't encrypt message properly")
	})

	t.Run("Decrypt", func(t *testing.T) {
		plain, err := decryptor.Decrypt(cipher)
		require.NoError(t, err)
		require.NotEmpty(t, plain, "Decryption attempt returned no data")
		require.Equal(t, msg, string(plain))
	})
}

// prepareTestGenerateKeysFiles writes keys to temporary test files.
func prepareTestGenerateKeysFiles(t *testing.T, privatePEM, publicPEM []byte) (pub, priv string) {
	t.Helper()

	// write keys in temporary directory
	dir := t.TempDir()

	// write private key to file
	fPriv, err := os.CreateTemp(dir, "test_genkey_*")
	require.NoError(t, err, "failed to create temporary private key file: %v", err)
	defer fPriv.Close()

	_, err = fPriv.Write(privatePEM)
	require.NoError(t, err, "failed to write private key to temporary file: %v", err)

	// write public key to file
	fPub, err := os.CreateTemp(dir, "test_genkey_pub*")
	require.NoError(t, err, "failed to create temporary public key file: %v", err)
	defer fPub.Close()

	_, err = fPub.Write(publicPEM)
	require.NoError(t, err, "failed to write public key to temporary file: %v", err)

	return fPub.Name(), fPriv.Name()
}
