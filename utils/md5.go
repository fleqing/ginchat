package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// Md5Encode
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Md5EncodeUpper
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// 加密
func MakePassword(data string, salt string) string {
	return Md5Encode(data + salt)
}

// 解密
func ValidPassword(data string, salt string, password string) bool {
	return Md5Encode(data+salt) == password
}
