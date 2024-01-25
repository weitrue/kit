package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"time"
)

func Withdraw(ctx context.Context, from, to string) {
	client := ethclient.NewClient(ETHClient)
	balanceAt, err := client.BalanceAt(ctx, common.HexToAddress(from), nil)
	if err != nil {
		return
	}

	gasTip := big.NewInt(1000000000) // 设置 tip（最小费用）
	//gasFeeCap := big.NewInt(5000000000) // 设置 fee cap（最大费用）
	gasLimit := uint64(21000)
	lowest := gasTip.Mul(gasTip, new(big.Int).SetUint64(gasLimit))
	if balanceAt.Cmp(lowest) <= 0 {
		return
	}

	tx, err := CreateTransaction(ctx, client, from, to, *balanceAt)
	if err != nil {
		return
	}

	tx, err = signTransaction(tx, "c29fc418e770e259ebd5e02e6393191898415eabffc5be64cfeaa47c1bddeeaf")
	if err != nil {
		return
	}

	err = client.SendTransaction(ctx, tx)
	if err != nil {
		return
	}

	receipt, err := waitTransactionReceipt(ctx, client, tx.Hash())
	if err != nil {
		return
	}

	fmt.Println(receipt.GasUsed)
}

func CreateTransaction(ctx context.Context, client *ethclient.Client, sender, to string, value big.Int) (*types.Transaction, error) {
	nonce, err := getNonce(ctx, client, common.HexToAddress(sender))
	if err != nil {
		return nil, err
	}

	//gasTip := big.NewInt(1000000000)    // 设置 tip（最小费用）
	//gasFeeCap := big.NewInt(5000000000) // 设置 fee cap（最大费用）
	gasLimit := uint64(21000)
	receiver := common.HexToAddress(to)
	var data []byte

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	// gas limit
	estimateGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: common.HexToAddress(sender),
		To:   &receiver,
		Data: data,
	})
	if gasLimit < estimateGas {
		gasLimit = estimateGas
	}

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &receiver,
		Value:    &value,
		Data:     data,
	}), nil
}

func signTransaction(tx *types.Transaction, privateKeyHex string) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return types.SignTx(tx, types.NewEIP155Signer(tx.ChainId()), privateKey)
}

func getNonce(ctx context.Context, client *ethclient.Client, address common.Address) (uint64, error) {
	return client.PendingNonceAt(ctx, address)
}

func waitTransactionReceipt(ctx context.Context, client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err == nil && receipt != nil {
			return receipt, nil
		}

		// 如果交易还未被打包，则等待一段时间再次尝试
		time.Sleep(5 * time.Second)
	}
}
