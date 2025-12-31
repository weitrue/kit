package tron

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/weitrue/kit/token"
)

const (
	NativeToken = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

var (
	erc20Contract, _ = token.NewErc20Filterer(common.Address{}, nil)
	wrapContract, _  = token.NewTokenFilterer(common.Address{}, nil)
	TronDecimal      = 6
	ErrNotFound      = fmt.Errorf("transaction info not found")
	ErrChecking      = fmt.Errorf("transaction is checking")
)

type TxSearchMeta struct {
	ChainId int64  `json:"chainId"`
	Hash    string `json:"hash"`
}

type TxTransferMeta struct {
	TxSearchMeta
	Index      int64            `json:"index"`
	From       string           `json:"from"`
	To         string           `json:"to"`
	FromLabel  string           `json:"fromLabel,omitempty"`
	ToLabel    string           `json:"toLabel,omitempty"`
	Assets     []TransferAsset  `json:"assets"`
	Amount     decimal.Decimal  `json:"amount"`
	Value      decimal.Decimal  `json:"value"`
	TxTime     int64            `json:"txTime"`
	IsTransfer bool             `json:"isTransfer"`
	Detail     TxDetail         `json:"detail"`
	Transfers  []TxTransferMeta `json:"transfers,omitempty"`
}

type TransferAsset struct {
	ChainId int64  `json:"chainId"`
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
	Logo    string `json:"logo"`
	Decimal int    `json:"decimal"`
	Native  bool   `json:"-"`
}

type TxDetail struct {
	Fee   *decimal.Decimal `json:"fee,omitempty"`
	Block int64            `json:"block,omitempty"`
	Type  *int64           `json:"type,omitempty"`
}
