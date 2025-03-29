package utils

import (
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)

func TestToDecimal(t *testing.T) {
	type args struct {
		val      any
		decimals int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				val:      decimal.Zero,
				decimals: 18,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, _ := decimal.NewFromString("173383212446124014651")
			fmt.Println(ToDecimal(val, tt.args.decimals))
		})
	}
}
