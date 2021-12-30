/*
 * @Descripation: aes加密
 * @Date: 2021-12-20 10:00:55
 */

package client

import (
	"crypto/aes"
	"crypto/cipher"
	"opensca/internal/logs"
)

// aes-tag大小
const tagSize = 16

/**
 * @description: aes-gcm加密
 * @param {[]byte} text 原文
 * @param {[]byte} key aes-key 16子节
 * @param {[]byte} nonce aes-nonce 16子节
 * @return {[]byte} 密文
 * @return {[]byte} aes-tag 16子节
 */
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

/**
 * @description: aes-gcm解密
 * @param {[]byte} ciphertext 密文
 * @param {[]byte} key aes-key 16子节
 * @param {[]byte} nonce aes-nonce 16子节
 * @param {[]byte} tag aes-tag 16子节
 * @return {[]byte} 原文
 */
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
