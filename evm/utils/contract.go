package utils

import "golang.org/x/crypto/sha3"

const (
	eip165SupportsInterfaceABI = ""
	nft721InterfaceId          = ""
)

// IsERC721 supportsInterface(0x80ac58cd)是否返回true
func IsERC721() {

}

func CalculateInterfaceId(contractABI string) []byte {
	// 计算函数签名的 keccak256 哈希值
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(contractABI))
	interfaceId := hash.Sum(nil)[:4] // 只取前4个字节作为 interfaceId
	return interfaceId
}

func calculateSelector(selector string) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(selector))
	return hash.Sum(nil)[:4]
}
