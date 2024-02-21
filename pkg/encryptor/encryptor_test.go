package encryptor

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEncryptor tests encrypt-decrypt scenario.
func TestEncryptor(t *testing.T) {
	// paths to files containing encryption keys
	pubPKCS1, priPKCS1 := prepareTestEncryptorKeys(t, "PKCS1")
	pubPKCS8, priPKCS8 := prepareTestEncryptorKeys(t, "PKCS8")

	msg := "Hello, World! Привет!"

	type args struct {
		private string
		public  string
	}
	tests := []struct {
		name string
		args args
		msg  string
	}{
		{
			name: "PKCS1",
			args: args{private: priPKCS1, public: pubPKCS1},
			msg:  msg,
		},
		{
			name: "PKCS8",
			args: args{private: priPKCS8, public: pubPKCS8},
			msg:  msg,
		},
		// also test longer inputs that will require chunked encryption
		{
			name: "PKCS1-x100",
			args: args{private: priPKCS1, public: pubPKCS1},
			msg:  strings.Repeat(msg, 100),
		},
		{
			name: "PKCS8-x100",
			args: args{private: priPKCS8, public: pubPKCS8},
			msg:  strings.Repeat(msg, 100),
		},
		{
			name: "PKCS1-x500",
			args: args{private: priPKCS1, public: pubPKCS1},
			msg:  strings.Repeat(msg, 500),
		},
		{
			name: "PKCS8-x500",
			args: args{private: priPKCS8, public: pubPKCS8},
			msg:  strings.Repeat(msg, 500),
		},
	}

	// encrypted message
	var cipher []byte

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encryptor, err := NewEncryptor(tc.args.public)
			require.NoError(t, err, "failed to create Encryptor")
			decryptor, err := NewDecryptor(tc.args.private)
			require.NoError(t, err, "failed to create Decryptor")

			t.Run("Encrypt", func(t *testing.T) {
				var err error
				cipher, err = encryptor.Encrypt([]byte(tc.msg))
				require.NoError(t, err)
				require.NotEmpty(t, cipher, "Encryption attempt returned no data")
				require.NotContains(t, string(cipher), tc.msg, "Encryptor didn't encrypt message properly")
			})

			t.Run("Decrypt", func(t *testing.T) {
				plain, err := decryptor.Decrypt(cipher)
				require.NoError(t, err)
				require.NotEmpty(t, plain, "Decryption attempt returned no data")
				require.Equal(t, tc.msg, string(plain))
			})
		})
	}
}

// prepareTestEncryptorKeys creates test files containing encryption keys.
func prepareTestEncryptorKeys(t *testing.T, form string) (pub, priv string) {
	t.Helper()

	var pubKey, privKey string

	switch form {
	case "PKCS1":
		pubKey = PKCS1TestPublicKey
		privKey = PKCS1TestPrivateKey
	case "PKCS8":
		pubKey = PKCS8TestPublicKey
		privKey = PKCS8TestPrivateKey
	default:
		require.FailNow(t, "fatal: unexpected key form provided for tests preparations: '%s'", form)
		return "", ""
	}

	// write keys in temporary directory
	dir := t.TempDir()

	// public key
	fPub, err := os.CreateTemp(dir, "test_enc_pub*")
	require.NoError(t, err, "failed to create temporary public key file: %v", err)
	defer fPub.Close()

	_, err = fPub.WriteString(pubKey)
	require.NoError(t, err, "failed to write public key to temporary file: %v", err)

	// private key
	fPriv, err := os.CreateTemp(dir, "test_enc_*")
	require.NoError(t, err, "failed to create temporary private key file: %v", err)
	defer fPriv.Close()

	_, err = fPriv.WriteString(privKey)
	require.NoError(t, err, "failed to write private key to temporary file: %v", err)

	return fPub.Name(), fPriv.Name()
}

// encryption keys for tests
const (
	PKCS1TestPublicKey = `
-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEApjXFJWp48sG+w5vUCKSaRjXXxqkecxbmEIBeSBfq9CpwakdE9hMp
uvBRMsfMtlQSAnW2UwvlhCpFwjpz7ofeEPeeiLaooXnrS+on3LNr6+fpPI+w6Fhw
QHbFIY4QA1QgfO2hxZPDfF4kRaumfojcuL6L1I0WwNE9aukz4ZicgGWWsS3jV6mj
RtmIsC1776HqIdzDc+j606jgyg0H2QCosEHi8+FiEoJzmvMxr7Wdy7VYmzP8GT5Z
N7YP4tUMbhTRvke8kNVOK7q76mq7No8zE3d9ukS2ifX8X+dEjrwNz8ZBfNlTJOEv
wggDP7E6OLTJN/4c8u/EEepEq/CYKKQ3nQIDAQAB
-----END RSA PUBLIC KEY-----
`

	PKCS1TestPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApjXFJWp48sG+w5vUCKSaRjXXxqkecxbmEIBeSBfq9CpwakdE
9hMpuvBRMsfMtlQSAnW2UwvlhCpFwjpz7ofeEPeeiLaooXnrS+on3LNr6+fpPI+w
6FhwQHbFIY4QA1QgfO2hxZPDfF4kRaumfojcuL6L1I0WwNE9aukz4ZicgGWWsS3j
V6mjRtmIsC1776HqIdzDc+j606jgyg0H2QCosEHi8+FiEoJzmvMxr7Wdy7VYmzP8
GT5ZN7YP4tUMbhTRvke8kNVOK7q76mq7No8zE3d9ukS2ifX8X+dEjrwNz8ZBfNlT
JOEvwggDP7E6OLTJN/4c8u/EEepEq/CYKKQ3nQIDAQABAoIBAH3/SvknMpLMXqGI
cn8u0KiiznUR5TxXwoYF2fMK9pirx1Y5usRUGJFW9ibpoX0iVBi7RUDFRvMe3Iz1
pMbRBn2USQDYfxMSClYdJqS++vP1dTDsuK4ZxNaJcr0SJX9wecRanATcFBgmgZaz
Hd7UP4ZpJDU302TN+aIfMK4284S8xWNpexaUuYbULXprY5jKNpPbkLSeDZk3Z5Ba
P6cbgTKRp8LEcKhCvFgSu+Sefp8LRqdMWj0gypAPkMeoBkqX1WxBKx3pFuOoXXGF
oDVZbIfT/5o4ZH+HWDQxQ0rbbU+g6vn8/zgSOnpL54mdpbAbX/VJReiT+HjJXlTO
8p1ou90CgYEA0sYLasX5m0cwl0KhKWyd2OAGikb2Y5TkYMiC1E9xUYq5L2R6fjM1
EhudA0mNkn0mUub5wCjXu7Tl9dpZfCVS+yZHJScHMLX0bnqnAofe57oDTct3f6BZ
M92/vDCTWww9Ep0G6/UUQT03PUnUEmmEiGvVxEs7YHJiAoGKtBDh6asCgYEAyd/Q
gGlUzd8THdJv3ObRWIQTh4OIWlTngS/xROPwx5uU+/IRcWXFCa2NaPG3A71bONPw
MJBu/DIiN/73YI4am1SAcqWqPzfEzW9BCRvMKrcIUeOesh+UTZAtmv1bxyNTnvnA
VLP4Lp9+wX18ktdJ2i4Cb4OWuUp9cnZtYD0769cCgYEAgUwCNvbJnyq1sSSrjqjV
z/Plry/G9+Gfk4uNTkaJolbyMRN3XTF2jewojpqxYTqqptKQRQQJC1n2c7IFkPWo
iO67WBwv+f60uo09JH1LDWX7nt6BKOapsfqHIx+9VW4VjPSNCXRnf3nZCBzhHoZi
Sfm0wdzQyOKCIz1qY/mzTE8CgYAOuVBPzV7uFRMj3bFi/0LNnfR+sc1EKWUpOwHx
8F4jcOmQ5rr29mFEr2/c86RRlYINxweBw0cVBeMRCnuogzTVl3g25HZiXgLwqOip
bWmaw/bLYjq06zC554YsA8ap4525vqWUh/vWCrQIEMsBOsqcKzbqQ0K/CPvVWWXo
2w5/yQKBgBEG8JJFDbO+7+l45HlOxZeFsUvU+cWm4RrTP3mkrvS8tQqRvC1VIRe0
NN1c1+5KFwmaC4VFfhH/5lir+rwUeu2lujte1pck7A8xoxpM6Dh6k30SIKFkB6BY
ubYR5g4gbiyjme4GwV5EE0HXaazfwKSznwz7byFwvd5yk6gPFb7+
-----END RSA PRIVATE KEY-----
`

	PKCS8TestPublicKey = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAopcaye8TZmJ/T2K7c7J3
9Zxw/3wRVPVe7PU5UURnekukI3gDLJMndL6PL5L7LekQ9k1s3UkkqJRK9qOlGMzM
A9dha1kX5eWE9TaWez4es1oHGEd8eggMjQmk8NffgpUxDsu71KPsuSojyMxQj049
Vd8sY79JEbmnyeYo6BxArFKojoV/IDB17EWKPsTmUYdRXNuNXYUSg1PERYqvTwG7
k+Hbv+W30lbRuw0hYX04lYTjG+9ERv7Y9k2Op2psVGdKPdZA1JZxgM21k7gIjsea
36GgEXqEjkGr+pBgQeLXJJ6HT2EfcPxRB2eTl5XgxCCqROPfUOjRee1KBnTkUAZV
JwIDAQAB
-----END PUBLIC KEY-----
`

	PKCS8TestPrivateKey = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCilxrJ7xNmYn9P
Yrtzsnf1nHD/fBFU9V7s9TlRRGd6S6QjeAMskyd0vo8vkvst6RD2TWzdSSSolEr2
o6UYzMwD12FrWRfl5YT1NpZ7Ph6zWgcYR3x6CAyNCaTw19+ClTEOy7vUo+y5KiPI
zFCPTj1V3yxjv0kRuafJ5ijoHECsUqiOhX8gMHXsRYo+xOZRh1Fc241dhRKDU8RF
iq9PAbuT4du/5bfSVtG7DSFhfTiVhOMb70RG/tj2TY6namxUZ0o91kDUlnGAzbWT
uAiOx5rfoaAReoSOQav6kGBB4tcknodPYR9w/FEHZ5OXleDEIKpE499Q6NF57UoG
dORQBlUnAgMBAAECggEADvNgG5N9WtvRzcX2wLkiKMFIRBCaM2VrY39afuuQVwGK
SAIAILguAkvc33NfiiWKpsC7MvnyggHo3BceK6f2q/UWgzLVF2cXO8ks5eDyHRc5
McquWouk7nnQOJ6mcdva1VEh+3f7YOd2v5RyD/YSKR71dlKfmx6m8XzVjspnnrDG
6pLk+GOu1tb+O4WZNTn5AuMZ2rqaxJ1Zx9HZNjXc4IF0sCTiu1cBRwFMS/JoQ8xU
flcV9uZR66t9NI13+iEFluKYNHBFpjRBNPz2Gl/gagGJUjdyrC6aSSwEjZzHHPBt
r7aH1bQmGfAM+fixJRj9qNToMQ3CgUMQ74pwLpuQsQKBgQDR7OY/yS/DrmdotlRV
qptVOyOFv72uQv4uSblV0TsG7IAOO3dLrts5ev8ptgbBi0LWkQvSaQYbKXGRL0Pk
yEUVr7CPcMUSdTE50Ou7SzUs3GjUL7Fmqya/B9X7vvy1ISuh9icMkyhFD3bST5v8
1+TWmeBJfClZYTlp/S/0a2MzbwKBgQDGRpSwcEAuKaA79kz+BIpiMHgJhWAvn4cH
vQw50A6mOSVYX7Q8otmFRnkTW0ItUGG6EvhyBgEW3hbCfaiROL6243fRSok40j0c
rgvgbfQ2yxRKwKktI5tWbAZJw7WwUfGEXDSVqtastcOS5+2jDN3tcmgB7kQ105Ak
EFG8Q3a9yQKBgQC13Z/oONr5kzmhXPyZLvHdmBuBPwkMVJuy7En1UAvsRq32JoyG
c9kW5jjzIPCfco+YJhbw1e3lUDVES4dtSBeZ3xh+XHtPL1pTNW1UWKab0+O5mAv/
31z0+MvoiqCSJ0eOzBsuaLrlga02LKP+G+f0B2FpKHkGJxW/fgqynOpXXwKBgHnf
SmfFJ334chhJRXvhmJnQBwh7Jes/2ETma8xhWY/MkmrmsKeQxblI3wUJ0/x8awVv
wax8ilDfBAhaKrHQulE6Mhy//uahO5UUWhKCI1lGrFtiXtpzB3kwfxD6LgSJ6bUc
4+mXD8srEmKLEqNlNkhMSSfep6iaOQAh08uvgB4ZAoGBAIKWpzsrQl3OK301gbT2
QW5ayXCVS1NaDxQUZaLiICu8vWVMWgrAC0f6msthCD76fMfwZSr/Bw6AwWV8+mOp
qjU+cZU7CJAXtQIYyCYrE3TTi5DQ/mHM9IarfcTZ6/bjdUSaa+QwQwP78XbQYcak
Kz3iTEvzUF+ulaf0OWCiKibv
-----END PRIVATE KEY-----
`
)
