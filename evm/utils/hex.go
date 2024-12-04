package utils

import "encoding/hex"

// RemoveZeroHex delete the 0x from the front
func RemoveZeroHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

func Hex2Bytes(str string) []byte {
	h, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return h
}
