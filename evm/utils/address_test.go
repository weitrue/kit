package utils

import (
	"context"
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"reflect"
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
