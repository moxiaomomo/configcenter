package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

const str_length = 80

// UniqueId_b64 uniqid with base64 format
func UniqueId_b64() string {
	b := make([]byte, str_length)
	rand.Read(b)
	en := base64.StdEncoding // or URLEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	//	fmt.Printf("src=%s\ndst=%s\n", b, d)
	return string(d)
}

// UniqueID refer to php uniqid
func UniqueID(prefixs ...string) string { // refer to php uniqid
	str := ""
	for _, p := range prefixs {
		if str == "" {
			str = p
		} else {
			str = fmt.Sprintf("%s_%s", str, p)
		}
	}
	tsStr := strconv.FormatInt(int64(time.Now().UnixNano()*1000000), 16)[4:]
	return fmt.Sprintf("%s_%s", str, tsStr)
}
