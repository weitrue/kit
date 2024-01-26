package utils

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Withdraw(ctx context.Context, from, to string) error {
	client := ethclient.NewClient(ETHClient)
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	balanceAt, err := client.BalanceAt(ctx, common.HexToAddress(from), nil)
	if err != nil {
		return err
	}

	gas := big.NewInt(2000000000) // 设置 tip（最小费用）
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	if gas.Cmp(gasPrice) < 0 {
		gas = gasPrice
	}

	lowest := new(big.Int).Mul(gas, big.NewInt(int64(gasLimit)))
	if balanceAt.Cmp(lowest) <= 0 {
		return errors.New("")
	}

	balanceAt.Sub(balanceAt, lowest)
	tx, err := CreateTransaction(ctx, client, from, to, balanceAt, gas, gasLimit, "0x")
	if err != nil {
		return err
	}

	tx, err = signTransaction(tx, "c29fc418e770e259ebd5e02e6393191898415eabffc5be64cfeaa47c1bddeeaf", chainID)
	if err != nil {
		return err
	}

	err = client.SendTransaction(ctx, tx)
	if err != nil {
		return err
	}

	receipt, err := waitTransactionReceipt(ctx, client, tx.Hash())
	if err != nil {
		return err
	}

	fmt.Println(receipt.GasUsed)
	return nil
}

func CreateTransaction(ctx context.Context, client *ethclient.Client, sender, to string, value, gas *big.Int, gasLimit uint64, input string) (*types.Transaction, error) {
	receiver := common.HexToAddress(to)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	if gas.Cmp(gasPrice) < 0 {
		gas = gasPrice
	}

	var data []byte
	if input != "0x" {
		data, err = hex.DecodeString(input[2:])
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

	nonce, err := getNonce(ctx, client, common.HexToAddress(sender))
	if err != nil {
		return nil, err
	}

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gas,
		Gas:      gasLimit,
		To:       &receiver,
		Value:    value,
		Data:     data,
	}), nil
}

func signTransaction(tx *types.Transaction, privateKeyHex string, chainID *big.Int) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	//publicKey := privateKey.Public()
	//publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//if !ok {
	//}
	//fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
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
