package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EnCrypt_DeCrypt_Succeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		str     string
		wantErr bool
	}{
		{"pingcap", false},
		{"pingcap123", false},
		{"admin123", false},
	}

	for _, tt := range tests {
		encrypt, err := AesEncryptCFB(tt.str)
		assert.Equal(t, tt.wantErr, err != nil)
		t.Logf("Encrypt %s -> %s\n", tt.str, encrypt)
		decrypt, err := AesDecryptCFB(encrypt)
		assert.Equal(t, tt.wantErr, err != nil)
		assert.Equal(t, tt.str, decrypt)
	}
}

func Test_DeCrypt_Succeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		plainStr  string
		cryptoStr string
		wantErr   bool
	}{
		{"pingcap", "4482c20f6a6358703287c47238a4df8aa9ee4b03db3873", false},
		{"pingcap123", "89d31e849f73c11d1921c69ea95aa2525b73ce0309f53ace51cd", false},
		{"admin2", "4bc5947d63aab7ad23cda5ca33df952e9678d7920428", false},
	}

	for _, tt := range tests {
		decrypt, err := AesDecryptCFB(tt.cryptoStr)
		t.Logf("Decrypt %s -> %s\n", tt.cryptoStr, decrypt)
		assert.Equal(t, tt.wantErr, err != nil)
		assert.Equal(t, tt.plainStr, decrypt)
	}
}

func Test_DeCrypt_Failed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		plainStr  string
		cryptoStr string
		wantErr   bool
	}{
		{"", "pingcapa6358703287c47238a4df8aa9ee4b03db3873", true}, // Inleagl hex code
		{"", "89d31e849f73c11d1921c69ea95aa2", true},               // length less than aes.BlockSize(16)
	}

	for _, tt := range tests {
		decrypt, err := AesDecryptCFB(tt.cryptoStr)
		t.Logf("Decrypt %s failed, %v\n", tt.cryptoStr, err)
		assert.Equal(t, tt.wantErr, err != nil)
		assert.Equal(t, "", decrypt)
	}
}
