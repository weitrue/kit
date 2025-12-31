package tron

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

func LogTronToEvm(logIndex uint, rawLog *core.TransactionInfo_Log) ethtypes.Log {
	log := ethtypes.Log{
		Address: ethcommon.BytesToAddress(rawLog.GetAddress()),
		Topics:  make([]ethcommon.Hash, 0, len(rawLog.Topics)),
		Data:    rawLog.GetData(),
		Index:   logIndex,
	}
	for _, topic := range rawLog.Topics {
		log.Topics = append(log.Topics, ethcommon.BytesToHash(topic))
	}
	return log
}
