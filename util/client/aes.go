/*
 * @Description: aes加密
 * @Date: 2021-12-20 10:00:55
 */

package client

import (
	"crypto/aes"
	"crypto/cipher"
	"util/logs"
)

// aes-tag大小
const tagSize = 16

// encrypt aes-gcm加密
func encrypt(text, key, nonce []byte) (ciphertext, tag []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logs.Error(err)
		return
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		logs.Error(err)
		return
	}
	res := aesgcm.Seal(nil, nonce, text, nil)
	tagIndex := len(res) - tagSize
	return res[:tagIndex], res[tagIndex:]
}

// decrypt aes-gcm解密
func decrypt(ciphertext, key, nonce, tag []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logs.Error(err)
		return
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		logs.Error(err)
		return
	}
	text, err = aesgcm.Open(nil, nonce, append(ciphertext, tag...), nil)
	if err != nil {
		logs.Error(err)
	}
	return
}
