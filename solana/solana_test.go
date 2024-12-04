package solana

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/btcsuite/btcutil/base58"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"

	"github.com/blocto/solana-go-sdk/client"
)

type TokenBalance struct {
	IsNative    bool   `json:"isNative"`
	Mint        string `json:"mint"`
	Owner       string `json:"owner"`
	State       string `json:"state"`
	TokenAmount struct {
		Amount         string  `json:"amount"`
		Decimals       int     `json:"decimals"`
		UiAmount       float64 `json:"uiAmount"`
		UiAmountString string  `json:"uiAmountString"`
	} `json:"tokenAmount"`
}

func Test_AddressHoldTokens(t *testing.T) {
	endpoint := "https://long-virulent-seed.solana-mainnet.quiknode.pro/b93562d54b8e0cc4d5f1aec08f0d0dbe293ccb1e/" // 替换为你使用的 Solana RPC 端点
	client := rpc.New(endpoint)

	// 设置要查询的钱包地址
	walletAddress, err := solana.PublicKeyFromBase58("5tzFkiKscXHK5ZXCGbXZxdw7gTjjD1mBwuoFbhUvuAi9")
	assert.Nil(t, err)

	res, err := client.GetTokenAccountsByOwner(
		context.Background(),
		walletAddress,
		&rpc.GetTokenAccountsConfig{
			ProgramId: &solana.TokenProgramID,
		},
		&rpc.GetTokenAccountsOpts{},
	)

	assert.Nil(t, err)
	if res != nil {

		for _, v := range res.Value {
			data := v.Account.Data.GetBinary()
			tb := TokenBalance{}
			_ = json.Unmarshal(data, &tb)
			fmt.Println(tb)
			dataS := base64.StdEncoding.EncodeToString(data)
			fmt.Println(dataS)

			var tokenAccount TokenAccount
			copy(tokenAccount.Mint[:], data[:32])
			copy(tokenAccount.Owner[:], data[32:64])
			tokenAccount.Amount = binary.LittleEndian.Uint64(data[64:72])
			fmt.Println("address:", v.Pubkey.String(), "tokenAddress:", base58.Encode(tokenAccount.Mint[:]), "owner:", base58.Encode(tokenAccount.Owner[:]), "amount: ", tokenAccount.Amount)

			//tokenMintAddress, _ := solana.PublicKeyFromBase58(base58.Encode(tokenAccount.Mint[:]))
			//accountInfo, err := client.GetAccountInfo(context.Background(), tokenMintAddress)
			//if err == nil {
			//	acc := accountInfo.GetBinary()
			//	fmt.Println(acc[44])
			//}
		}
	}
}

type TokenAccount struct {
	Mint   [32]byte
	Owner  [32]byte
	Amount uint64
}

func Test_AddressHoldTokensV2(t *testing.T) {
	endpoint := "https://long-virulent-seed.solana-mainnet.quiknode.pro/b93562d54b8e0cc4d5f1aec08f0d0dbe293ccb1e/" // 替换为你使用的 Solana RPC 端点
	client := client.NewClient(endpoint)

	res, err := client.GetTokenAccountsByOwnerByProgram(
		context.Background(),
		"5tzFkiKscXHK5ZXCGbXZxdw7gTjjD1mBwuoFbhUvuAi9",
		common.TokenProgramID.String(),
	)

	assert.Nil(t, err)
	for _, v := range res {
		fmt.Println("address:", v.PublicKey.String(), "tokenAddress:", v.Mint.String(), "owner:", v.Owner.String(), "amount: ", v.Amount)
	}
}

func Test_BlockHash(t *testing.T) {
	endpoint := "https://long-virulent-seed.solana-mainnet.quiknode.pro/b93562d54b8e0cc4d5f1aec08f0d0dbe293ccb1e/" // 替换为你使用的 Solana RPC 端点
	c := client.NewClient(endpoint)

	res, err := c.GetLatestBlockhash(
		context.Background(),
	)

	assert.Nil(t, err)
	fmt.Println(res.Blockhash)
	fmt.Println(res.LatestValidBlockHeight)

}

func Test_Balance(t *testing.T) {
	endpoint := "https://long-virulent-seed.solana-mainnet.quiknode.pro/b93562d54b8e0cc4d5f1aec08f0d0dbe293ccb1e/" // 替换为你使用的 Solana RPC 端点
	c := client.NewClient(endpoint)
	balance, err := c.GetBalance(context.Background(), "5yEnvhM4Ld3UZs2n173J2iR369E1ddcbQYeLSZxk4cYj")
	assert.Nil(t, err)
	fmt.Println(balance)
}

func TestName(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)
	fmt.Println(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()).Unix())
}
