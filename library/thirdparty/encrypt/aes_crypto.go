package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// key should be 16、24 or 32 length [] byte, conresponding to AES-128, AES-192 或 AES-256.
var key []byte

func init() {
	key = []byte(">]t1emf0rp1nGcap$t!Em@p!ngcap;[<")
}

func AesEncryptCFB(plain []byte) (encrypted []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "newCipher err, %s", err)
	}
	encrypted = make([]byte, aes.BlockSize+len(plain))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, status.Errorf(codes.Internal, "init vector err, %s", err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], plain)
	return encrypted, nil
}

func AesDecryptCFB(encrypted []byte) (decrypted []byte, err error) {
	block, _ := aes.NewCipher(key)
	if len(encrypted) < aes.BlockSize {
		return nil, status.Errorf(codes.Internal, "ciphertext too short, %d < aes.BlockSize(%d)", len(encrypted), aes.BlockSize)
	}
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted, nil
}
