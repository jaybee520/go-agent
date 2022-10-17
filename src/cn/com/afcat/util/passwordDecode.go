package util

import (
	"encoding/base64"
	"log"
	"strings"
)

func PasswdDecode(password string) string {
	sDec, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		log.Printf("密码解密失败: %s ", err.Error())
		panic("密码解密失败")
	}
	sDec1, err := base64.StdEncoding.DecodeString(string(sDec))
	if err != nil {
		log.Printf("密码解密失败: %s ", err.Error())
		panic("密码解密失败")
	}
	decodePass := strings.ReplaceAll(string(sDec1), "YLWS", "")
	return string(decodePass)
}
