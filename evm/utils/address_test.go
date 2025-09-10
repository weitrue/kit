package utils

import (
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mr-tron/base58"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyBytesToAddress(t *testing.T) {
	type args struct {
		publicKey []byte
	}
	tests := []struct {
		name string
		args args
		want common.Address
	}{
		{
			name: "",
			args: args{publicKey: make([]byte, 0)},
			want: common.HexToAddress(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PublicKeyBytesToAddress(tt.args.publicKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PublicKeyBytesToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	tx := new(types.Transaction)
	rawTxBytes, err := hex.DecodeString("f8651b85012a05f20082520894b8846e7acaa5f54f9916f188629bfc25f333d6c280808193a046d0a2d24e5dc473d2e3fbd9ee544e1f01799e8c960cd40f4655f7301e3a8909a023aaecab9631b341108681911b03156aa1366cdfa6efd9cbb454dc8f00041dec")
	assert.Nil(t, err)
	rlp.DecodeBytes(rawTxBytes, &tx)
	fmt.Println(tx.Hash().Hex())

	signer := types.NewEIP155Signer(tx.ChainId())
	sender, err := signer.Sender(tx)
	assert.Nil(t, err)

	fmt.Println(sender.Hex())
	md5 := crypto.MD5.New()
	md5.Write([]byte("Ruby@blocksec.com"))
	fmt.Println(hex.EncodeToString(md5.Sum(nil)))

	fmt.Println(len("bc1pc72wfxt28739kr0pa64nde09rzgwe8sw792nmxgu3pxwdfj8yymqmmhjl3"))
	fmt.Println(len("NativeLoader1111111111111111111111111111111"))

}

func TestGetBalance(t *testing.T) {
	type args struct {
		ctx     context.Context
		c       *rpc.Client
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "test",
			args: args{
				ctx:     context.Background(),
				c:       MerlinClient,
				address: "0x94661b622425271e8ce9cea52214bed9b9f18861",
			},
			want:    "0x0",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBalance(tt.args.ctx, tt.args.c, tt.args.address)
			assert.Nil(t, err)
			fmt.Println(got)
		})
	}
}

func TestAddress(t *testing.T) {
	fmt.Println(ethAddrToTronBase58(common.HexToAddress("0x659501f607c817ddc755c1c5b54b2bee1111e32b")))
	fmt.Println(TronToEth("TGrGhr66SrUbVRMNEkEgGqTQkDSno7ghUo"))
}

func ethAddrToTronBase58(a common.Address) string {
	// 1) 20 字节地址 + 0x41 版本字节
	payload := append([]byte{0x41}, a.Bytes()...) // 共 21 字节
	// 2) double SHA256
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	// 3) 追加前 4 字节校验和
	withChecksum := append(payload, h2[0:4]...)
	// 4) Base58 编码
	return base58.Encode(withChecksum)
}

// 把 TRON 的 hex 地址（带 41 前缀，21 字节）转成 Base58 (T...)
func tronHexToBase58(hexAddr string) (string, error) {
	s := strings.TrimPrefix(hexAddr, "0x")
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	if len(b) != 21 || b[0] != 0x41 {
		return "", fmt.Errorf("invalid tron hex address: %s", hexAddr)
	}
	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])
	withChecksum := append(b, h2[0:4]...)
	return base58.Encode(withChecksum), nil
}

func TronToEth(tronAddr string) (string, error) {
	// Base58Check 解码
	addrBytes, err := base58.Decode(tronAddr)
	if err != nil {
		return "", err
	}

	if len(addrBytes) != 25 {
		return "", fmt.Errorf("invalid tron address length: %d", len(addrBytes))
	}

	// 前21字节是真正的payload (0x41 + 20字节地址)，后4字节是checksum
	payload := addrBytes[:21]

	// 检查前缀是否是 0x41
	if payload[0] != 0x41 {
		return "", fmt.Errorf("invalid tron address prefix: 0x%x", payload[0])
	}

	// 取后20字节就是 ETH 的 hex 地址
	ethAddr := payload[1:]

	return "0x" + hex.EncodeToString(ethAddr), nil
}
