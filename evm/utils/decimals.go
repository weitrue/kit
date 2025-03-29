package utils

import (
	"math/big"

	"github.com/shopspring/decimal"
)

// ToDecimal wei to decimals
func ToDecimal(val any, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := val.(type) {
	case int64:
		value.SetInt64(v)
	case uint64:
		value.SetUint64(v)
	case string:
		value.SetString(v, 10)
	case decimal.Decimal:
		value = v.BigInt()
	case *big.Int:
		value = v
	}

	mul := decimal.New(1, int32(decimals))
	num := decimal.RequireFromString(value.String())
	result := num.DivRound(mul, int32(decimals))

	return result
}

// ToWei decimals to wei
func ToWei(value any, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := value.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromInt(v)
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

// CalcGasCost calculate gas cost given gas limit (units) and gas price (wei)
func CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return new(big.Int).Mul(gasLimitBig, gasPrice)
}
