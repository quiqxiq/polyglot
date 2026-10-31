package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBytesKey = "01234567890123456789012345678901"

func TestEncryptBytesDecryptBytes_RoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"text", []byte("hostname polyglot\r\n/ip firewall filter print")},
		{"binary", []byte{0x00, 0x01, 0xfe, 0xff, 0x00, 0x42, 0x7f}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := EncryptBytes(tc.plaintext, testBytesKey)
			require.NoError(t, err)
			// nonce 12 byte + GCM tag 16 byte di atas panjang plaintext.
			assert.Len(t, ciphertext, len(tc.plaintext)+12+16)
			assert.False(t, bytes.Equal(ciphertext, tc.plaintext))

			got, err := DecryptBytes(ciphertext, testBytesKey)
			require.NoError(t, err)
			assert.True(t, bytes.Equal(tc.plaintext, got), "plaintext round-trip")
		})
	}
}

func TestEncryptBytes_UsesRandomNonce(t *testing.T) {
	first, err := EncryptBytes([]byte("same plaintext"), testBytesKey)
	require.NoError(t, err)
	second, err := EncryptBytes([]byte("same plaintext"), testBytesKey)
	require.NoError(t, err)
	assert.NotEqual(t, first, second, "nonce acak harus menghasilkan ciphertext berbeda")
}

func TestEncryptBytes_CompatibleWithStringDecrypt(t *testing.T) {
	plaintext := "ssid=Polyglot"
	ciphertext, err := EncryptBytes([]byte(plaintext), testBytesKey)
	require.NoError(t, err)

	got, err := Decrypt(base64.StdEncoding.EncodeToString(ciphertext), testBytesKey)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestEncryptBytes_InvalidKeyLength(t *testing.T) {
	shortKey := "short-key"

	_, err := EncryptBytes([]byte("secret"), shortKey)
	assert.ErrorIs(t, err, ErrInvalidKeyLength)

	_, err = DecryptBytes([]byte("ciphertext"), shortKey)
	assert.ErrorIs(t, err, ErrInvalidKeyLength)
}

func TestDecryptBytes_TooShort(t *testing.T) {
	_, err := DecryptBytes([]byte{0x01, 0x02, 0x03}, testBytesKey)
	assert.ErrorIs(t, err, ErrCiphertextTooShort)
}

func TestDecryptBytes_WrongKey(t *testing.T) {
	ciphertext, err := EncryptBytes([]byte("secret"), testBytesKey)
	require.NoError(t, err)

	wrongKey := "abcdefghijklmnopqrstuvwxyz012345" // 32 byte, key berbeda
	_, err = DecryptBytes(ciphertext, wrongKey)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidKeyLength)
}

func TestDecryptBytes_TamperedCiphertext(t *testing.T) {
	ciphertext, err := EncryptBytes([]byte("integrity matters"), testBytesKey)
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)

	tampered := append([]byte(nil), ciphertext...)
	tampered[len(tampered)-1] ^= 0xff
	_, err = DecryptBytes(tampered, testBytesKey)
	require.Error(t, err, "GCM harus mendeteksi ciphertext yang diubah")
}
