package hash

import "testing"

func TestGetHash(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				str: "hello world",
			},
			want: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHash(tt.args.str); got != tt.want {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
