package gemini

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_tryGemini(t *testing.T) {
	type args struct {
		projectID string
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr error
	}{
		{
			name: "",
			args: args{
				projectID: "blocksec",
			},
			wantW:   "",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := tryGemini(w, tt.args.projectID)
			assert.Nil(t, err)
		})
	}
}
