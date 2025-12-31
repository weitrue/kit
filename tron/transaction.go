package tron

import (
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/shopspring/decimal"
	"github.com/weitrue/kit/evm/utils"
	"math/big"
	"strconv"
)

func GetTransactionTransfers(grpcClient *client.GrpcClient, hash string) (transaction TxTransferMeta, err error) {
	transfers := make([]TxTransferMeta, 0)
	txInfo, err := grpcClient.GetTransactionInfoByID(hash)
	if err != nil {
		if err.Error() == "transaction info not found" {
			return transaction, ErrNotFound
		}
		return transaction, err
	}

	contractRet := txInfo.GetReceipt().GetResult()
	if contractRet == core.Transaction_Result_DEFAULT {
		contractRet = core.Transaction_Result_SUCCESS
	}
	// ret checking
	if txInfo.GetResult() != core.TransactionInfo_SUCESS || contractRet != core.Transaction_Result_SUCCESS {
		return transaction, ErrChecking
	}

	tx, err := grpcClient.GetTransactionByID(hash)
	if err != nil {
		return transaction, err
	}

	fee := utils.ToDecimal(txInfo.GetFee(), TronDecimal)
	transaction = TxTransferMeta{
		TxSearchMeta: TxSearchMeta{
			Hash:    BytesToTronHash(txInfo.GetId()),
			ChainId: -2,
		},
		TxTime: txInfo.BlockTimeStamp,
		Detail: TxDetail{
			Block: txInfo.GetBlockNumber(),
			Fee:   &fee,
		},
	}
	contract := tx.GetRawData().Contract[0]
	contractType := contract.GetType()
	switch contractType {
	case core.Transaction_Contract_TriggerSmartContract: // 98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47
		msg := core.TriggerSmartContract{}
		err = contract.GetParameter().UnmarshalTo(&msg)
		if err != nil {
			return transaction, fmt.Errorf("parse TriggerSmartContract failed: %s", err.Error())
		}

		transaction.From = BytesToTronAddress(msg.GetOwnerAddress())
		transaction.To = BytesToTronAddress(msg.GetContractAddress())
		if msg.GetCallValue() != 0 {
			asset := TransferAsset{
				ChainId: -2,
				Address: NativeToken,
				Symbol:  "TRX",
				Decimal: TronDecimal,
			}
			trans := TxTransferMeta{
				TxSearchMeta: transaction.TxSearchMeta,
				From:         transaction.From,
				To:           transaction.To,
				Assets:       []TransferAsset{asset},
				Amount:       decimal.NewFromInt(msg.GetCallValue()),
				TxTime:       transaction.TxTime,
				Index:        int64(len(transfers)),
				IsTransfer:   true,
				Detail:       transaction.Detail,
			}
			transfers = append(transfers, trans)
		}

		transfers = ParseSmartContractTx(transaction.TxTime, txInfo, transfers)
	case core.Transaction_Contract_TransferContract: // a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65
		msg := core.TransferContract{}
		err = contract.GetParameter().UnmarshalTo(&msg)
		if err != nil {
			return transaction, fmt.Errorf("parse TransferContract failed: %s", err.Error())
		}
		transaction.From = BytesToTronAddress(msg.GetOwnerAddress())
		transaction.To = BytesToTronAddress(msg.GetToAddress())
		if msg.GetAmount() != 0 {
			asset := TransferAsset{
				ChainId: -2,
				Address: NativeToken,
				Symbol:  "TRX",
				Decimal: TronDecimal,
			}
			trans := TxTransferMeta{
				TxSearchMeta: transaction.TxSearchMeta,
				From:         transaction.From,
				To:           transaction.To,
				Assets:       []TransferAsset{asset},
				Amount:       decimal.NewFromInt(msg.GetAmount()),
				TxTime:       transaction.TxTime,
				Index:        int64(len(transfers)),
				IsTransfer:   true,
			}
			transfers = append(transfers, trans)
		}
	case core.Transaction_Contract_TransferAssetContract: // 2792c7e0c58aa4b2f0fcdd63f6fda9638e5052a1162fc2ea44cf863d6c80ca32
		msg := core.TransferAssetContract{}
		err = contract.GetParameter().UnmarshalTo(&msg)
		if err != nil {
			return transaction, fmt.Errorf("parse TransferAssetContract failed: %s", err.Error())
		}

		transaction.From = BytesToTronAddress(msg.GetOwnerAddress())
		transaction.To = BytesToTronAddress(msg.GetToAddress())
		if msg.GetAmount() != 0 {
			asset := TransferAsset{
				ChainId: -2,
				Address: string(msg.GetAssetName()),
			}

			tokenInfo, _ := grpcClient.GetAssetIssueByID(asset.Address)
			if tokenInfo != nil {
				asset.Symbol = string(tokenInfo.GetName())
				asset.Decimal = int(tokenInfo.GetPrecision())
			}
			trans := TxTransferMeta{
				TxSearchMeta: transaction.TxSearchMeta,
				From:         transaction.From,
				To:           transaction.To,
				Assets:       []TransferAsset{asset},
				Amount:       decimal.NewFromInt(msg.GetAmount()),
				TxTime:       transaction.TxTime,
				Index:        int64(len(transfers)),
				IsTransfer:   true,
			}
			transfers = append(transfers, trans)
		}
	}

	if len(transfers) == 0 {
		return transaction, nil
	}

	if len(transfers) == 1 && transfers[0].From == transaction.From && transfers[0].To == transaction.To {
		transfers[0].Index = 0
		transfers[0].Detail = transaction.Detail
		return transfers[0], nil
	} else {
		for i := 0; i < len(transfers); i++ {
			contain := false
			for _, asset := range transaction.Assets {
				if transfers[i].Assets[0].Address == asset.Address {
					contain = true
					break
				}
			}
			if !contain {
				transaction.Assets = append(transaction.Assets, transfers[i].Assets[0])
			}
		}
		transfers = append([]TxTransferMeta{transaction}, transfers...)
		transaction.Transfers = transfers
	}

	return transaction, nil
}

func ParseSmartContractTx(blockTimestamp int64, txInfo *core.TransactionInfo, transfers []TxTransferMeta) []TxTransferMeta {
	hash := BytesToTronHash(txInfo.GetId())
	logs := txInfo.GetLog()
	var offset uint16 = 1
	for logIndex, rawLog := range logs {
		ethLog := LogTronToEvm(uint(logIndex), rawLog)
		switch len(ethLog.Topics) {
		case 2:
			deposit, err := wrapContract.ParseDeposit(ethLog)
			if err == nil {
				if deposit.Wad != nil && deposit.Wad.Uint64() != 0 {
					asset := TransferAsset{
						ChainId: -2,
						Address: BytesToTronAddress(ethLog.Address.Bytes()),
					}
					tx := TxTransferMeta{
						TxSearchMeta: TxSearchMeta{
							Hash:    hash,
							ChainId: -2,
						},
						From:   BytesToTronAddress(ethLog.Address.Bytes()),
						To:     BytesToTronAddress(deposit.Dst.Bytes()),
						Assets: []TransferAsset{asset},
						Amount: decimal.NewFromInt(deposit.Wad.Int64()),
						TxTime: blockTimestamp,
						Index:  int64(len(transfers) + 1),
					}
					transfers = append(transfers, tx)
				}
			} else {
				withdrawal, err := wrapContract.ParseWithdrawal(ethLog)
				if err == nil {
					if withdrawal.Wad != nil {
						asset := TransferAsset{
							ChainId: -2,
							Address: BytesToTronAddress(ethLog.Address.Bytes()),
						}
						tx := TxTransferMeta{
							TxSearchMeta: TxSearchMeta{
								Hash:    hash,
								ChainId: -2,
							},
							From:       BytesToTronAddress(withdrawal.Src.Bytes()),
							To:         BytesToTronAddress(ethLog.Address.Bytes()),
							Assets:     []TransferAsset{asset},
							Amount:     decimal.NewFromInt(withdrawal.Wad.Int64()),
							TxTime:     blockTimestamp,
							Index:      int64(len(transfers) + 1),
							IsTransfer: true,
						}
						transfers = append(transfers, tx)
					}
				}
			}
		case 3:
			transfer, err := erc20Contract.ParseTransfer(ethLog)
			if err != nil {
				continue
			}
			if transfer.Value != nil && transfer.Value.Uint64() != 0 {
				asset := TransferAsset{
					ChainId: -2,
					Address: BytesToTronAddress(ethLog.Address.Bytes()),
				}
				tx := TxTransferMeta{
					TxSearchMeta: TxSearchMeta{
						Hash:    hash,
						ChainId: -2,
					},
					From:       BytesToTronAddress(transfer.From.Bytes()),
					To:         BytesToTronAddress(transfer.To.Bytes()),
					Assets:     []TransferAsset{asset},
					Amount:     decimal.NewFromInt(transfer.Value.Int64()),
					TxTime:     blockTimestamp,
					Index:      int64(len(transfers) + 1),
					IsTransfer: true,
				}
				transfers = append(transfers, tx)
			}
		case 4:
			// 721 transfer
		default:
		}
	}

	offset = 1 + uint16(len(logs))
	internalTxns := txInfo.GetInternalTransactions()
	for _, internalTxn := range internalTxns {
		//check internalTxn exec result
		if internalTxn.GetRejected() {
			continue
		}

		for _, callValue := range internalTxn.GetCallValueInfo() {
			if callValue.GetCallValue() == 0 {
				continue
			}
			asset := TransferAsset{
				ChainId: -2,
			}
			if callValue.GetTokenId() == "" {
				asset.Address = NativeToken
				asset.Symbol = "TRX"
				asset.Decimal = TronDecimal
			} else {
				tokenId, _ := strconv.ParseInt(callValue.GetTokenId(), 10, 64)
				asset.Address = address.Address(big.NewInt(tokenId).Bytes()).String()
			}
			tx := TxTransferMeta{
				TxSearchMeta: TxSearchMeta{
					Hash:    hash,
					ChainId: -2,
				},
				From:       BytesToTronAddress(internalTxn.GetCallerAddress()),
				To:         BytesToTronAddress(internalTxn.GetTransferToAddress()),
				Assets:     []TransferAsset{asset},
				Amount:     decimal.NewFromInt(callValue.GetCallValue()),
				TxTime:     blockTimestamp,
				Index:      int64(len(transfers) + 1),
				IsTransfer: true,
			}
			transfers = append(transfers, tx)
			offset = offset + 1
		}
	}

	return transfers
}

func BytesToTronHash(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	return common.Hash(hash).String()[2:]
}

func BytesToTronAddress(addr20 []byte) string {
	if len(addr20) == 21 && addr20[0] == 65 {
		return address.Address(addr20).String()
	}
	if len(addr20) != 20 {
		return ""
	}
	addr := append([]byte{0x41}, addr20...)
	return address.Address(addr).String()
}
