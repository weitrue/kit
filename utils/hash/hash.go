package hash

import (
	"crypto/md5"
	"encoding/hex"
)

func GetHash(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
