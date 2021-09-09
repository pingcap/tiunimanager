package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"io"

	utils "github.com/pingcap-inc/tiem/library/util/encrypt"
	"github.com/pingcap/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// key should be 16、24 or 32 length [] byte, conresponding to AES-128, AES-192 或 AES-256.
var key []byte

func init() {
	key = []byte(">]t1emf0rp1nGcap$t!Em@p!ngcap;[<")
}

func aesEncryptCFB(plain []byte) (encrypted []byte, err error) {
	encrypted = make([]byte, aes.BlockSize+len(plain))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, status.Errorf(codes.Internal, "init vector err, %s", err)
	}
	crypted, err := utils.AESEncryptWithCFB(plain, key, iv)
	if err != nil {
		return nil, err
	}
	cnt := copy(encrypted[aes.BlockSize:], crypted)
	if cnt != len(plain) {
		return nil, errors.Errorf("copy crypt to encrypted failed, expect %d, but %d", len(plain), cnt)
	}
	return encrypted, nil
}

func aesDecryptCFB(encrypted []byte) (decrypted []byte, err error) {
	if len(encrypted) < aes.BlockSize {
		return nil, status.Errorf(codes.Internal, "ciphertext too short, %d < aes.BlockSize(%d)", len(encrypted), aes.BlockSize)
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	return utils.AESDecryptWithCFB(encrypted, key, iv)
}

func AesEncryptCFB(plainStr string) (encryptedStr string, err error) {
	encrypted, err := aesEncryptCFB([]byte(plainStr))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), err
}

func AesDecryptCFB(encryptedStr string) (decryptedStr string, err error) {
	encrypted, err := hex.DecodeString(encryptedStr)
	if err != nil {
		return "", err
	}
	decrypted, err := aesDecryptCFB(encrypted)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
