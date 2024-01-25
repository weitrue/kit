package utils

import (
	"context"
	"testing"
)

func TestWithdraw(t *testing.T) {
	type args struct {
		ctx  context.Context
		from string
		to   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				ctx:  context.Background(),
				from: "0xba46dd807DD7A5bBe2eE80b6D0516A088223C574",
				to:   "0xEF87e7024Fe8f2D35fA8Be569a3c788722b2905f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Withdraw(tt.args.ctx, tt.args.from, tt.args.to)
		})
	}

}
