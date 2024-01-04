package xzap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetUp(t *testing.T) {
	type args struct {
		c    Config
		opts []Option
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				c: Config{
					ServiceName: "log",
					Mode:        "console",
					Path:        "logs/cron",
					Level:       "info",
					Compress:    false,
					KeepDays:    7,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetUp(tt.args.c, tt.args.opts...)
			assert.Nil(t, err)
			got.Info("test info")
			got.Error("test err")
		})
	}
}
