package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// md5 不可逆
// 小写
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	tempStr := h.Sum(nil)
	return hex.EncodeToString(tempStr)
}

// 大写
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// 加密
func MakePassword(plainpwd, salt string) string {
	// 将传过来的随机数 加上 一个随机数  再做一次 md5 加密
	return Md5Encode(plainpwd + salt)
}

// 解密
func ValidPassword(plainpwd, salt string, password string) bool {
	return Md5Encode(plainpwd+salt) == password
}
