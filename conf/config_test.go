package conf

import (
	"crypto/md5"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMustLoad(t *testing.T) {
	text := `{
	"a": "foo",
	"b": 1,
	"c": "${FOO}",
	"d": "abcd!@#$112"
}`
	t.Setenv("FOO", "2")
	type args struct {
		path string
		opts []Option
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: ".json",
			args: args{
				path: ".json",
			},
		},
		{
			name: ".yaml",
			args: args{
				path: ".yaml",
			},
		},
		{
			name: ".yml",
			args: args{
				path: ".yml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := createTempFile(tt.args.path, text)
			assert.Nil(t, err)
			defer os.Remove(tmpfile)

			var val struct {
				A string `json:"a"`
				B int    `json:"b"`
				C string `json:"c"`
				D string `json:"d"`
			}
			MustLoad(tmpfile, &val)
			assert.Equal(t, "foo", val.A)
			assert.Equal(t, 1, val.B)
			assert.Equal(t, "${FOO}", val.C)
			assert.Equal(t, "abcd!@#$112", val.D)
		})
	}
}

func createTempFile(ext, text string) (string, error) {
	tmpFile, err := os.CreateTemp(os.TempDir(), md5Hash([]byte(text))+"*"+ext)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tmpFile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return "", err
	}

	filename := tmpFile.Name()
	if err = tmpFile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}

func md5Hash(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}

func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}
