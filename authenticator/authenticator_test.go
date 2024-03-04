package authenticator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey("BlockSec", "weitrue")
	assert.Nil(t, err)
	t.Log(key)
}

func TestGenerateSecret(t *testing.T) {
	key, err := GenerateSecret("BlockSec", "weitrue")
	assert.Nil(t, err)
	t.Log(key)
}

func TestGenerateUrlAndSecret(t *testing.T) {
	url, secret, err := GenerateUrlAndSecret("BlockSec", "weitrue")
	assert.Nil(t, err)
	t.Log(url)
	t.Log(secret)
}

func TestGenerateCode(t *testing.T) {
	type args struct {
		domain   string
		userName string
		size     int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "test",
			args: args{
				domain:   "BlockSec",
				userName: "weitrue",
				size:     256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, uri, err := GenerateCode(tt.args.domain, tt.args.userName, tt.args.size)
			assert.Nil(t, err)
			t.Log(uri)

			os.WriteFile("qrcode_test.png", data, os.FileMode(0644))
		})
	}
}

func TestValidator(t *testing.T) {
	// code 可以从当前目录 qrcode.png 获得
	valid, err := Validate("otpauth://totp/BlockSec:weitrue?algorithm=SHA1&digits=6&issuer=BlockSec&period=30&secret=3BOK66HIXM27DFDCU3K7ISTE3R2DHISU", "208822")
	assert.Nil(t, err)
	if valid {
		fmt.Println("Validate Success!")
	} else {
		fmt.Println("Validate Failed!")
	}
}

func TestGenerateCodeAndSecret(t *testing.T) {
	type args struct {
		domain   string
		userName string
		size     int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "test",
			args: args{
				domain:   "BlockSec",
				userName: "weitrue",
				size:     256,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, uri, err := GenerateCodeAndSecret(tt.args.domain, tt.args.userName, tt.args.size)
			assert.Nil(t, err)
			t.Log(uri)

			os.WriteFile("qrcode_test.png", data, os.FileMode(0644))
		})
	}
}

func TestValidateCode(t *testing.T) {
	// code 可以从当前目录 qrcode.png 获得
	valid, err := ValidateCode("RVVFRB7H5FEYRQZQUEU37XQ6B74A465S", "428248")
	assert.Nil(t, err)
	if valid {
		fmt.Println("Validate Success!")
	} else {
		fmt.Println("Validate Failed!")
	}
}
