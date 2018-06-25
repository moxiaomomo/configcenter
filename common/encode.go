package common

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"regexp"
)

func Md5(s string) string {
	a := md5.Sum([]byte(s))
	return hex.EncodeToString(a[:])
}

func Base64Enc_std(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func Base64Enc_url(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

func Base64Dec_std(s string) string {
	uDec, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(uDec)
}

func Base64Dec_url(s string) string {
	uDec, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(uDec)
}

func IsFileHashValid(hash string) bool {
	if len(hash) != 40 {
		return false
	}
	match, _ := regexp.Compile("^[0-9A-Fa-f]{40}$")
	if !match.MatchString(hash) {
		return false
	}
	return true
}
