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
				val:      decimal.NewFromFloat(1000000000000000000),
				decimals: 18,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(ToDecimal(tt.args.val, tt.args.decimals))
		})
	}
}
