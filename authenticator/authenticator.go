package authenticator

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// GenerateKey 生成2FA qrcode text
func GenerateKey(domain, userName string) (string, error) {
	key, err := generateKey(domain, userName)
	if err != nil {
		return "", err
	}

	return key.URL(), nil
}

// GenerateSecret 生成2FA secret
func GenerateSecret(domain, userName string) (string, error) {
	key, err := generateKey(domain, userName)
	if err != nil {
		return "", err
	}

	return key.Secret(), nil
}

// GenerateUrlAndSecret 生成2FA qrcode text 和 secret
func GenerateUrlAndSecret(domain, userName string) (string, string, error) {
	key, err := generateKey(domain, userName)
	if err != nil {
		return "", "", err
	}

	return key.URL(), key.Secret(), nil
}

// GenerateCode 生成2FA 二维码图片和 qrcode text
func GenerateCode(domain, userName string, size int) ([]byte, string, error) {
	key, err := generateKey(domain, userName)
	if err != nil {
		return nil, "", err
	}

	code, err := generateCode(key.URL(), size)
	if err != nil {
		return nil, "", err
	}

	return code, key.URL(), nil
}

// GenerateCodeAndSecret 生成2FA 二维码图片和 secret
func GenerateCodeAndSecret(domain, userName string, size int) ([]byte, string, error) {
	key, err := generateKey(domain, userName)
	if err != nil {
		return nil, "", err
	}

	code, err := generateCode(key.URL(), size)
	if err != nil {
		return nil, "", err
	}

	return code, key.Secret(), nil
}

// Validate 根据url 验证 code
func Validate(url, code string) (bool, error) {
	key, err := otp.NewKeyFromURL(url)
	if err != nil {
		return false, err
	}

	return totp.Validate(code, key.Secret()), nil
}

// ValidateCode 根据secret验证code
func ValidateCode(secret, code string) (bool, error) {
	return totp.Validate(code, secret), nil
}

func generateKey(domain, userName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      domain,
		AccountName: userName,
	})
}

func generateCode(text string, size int) ([]byte, error) {
	qrCode, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	return qrCode.PNG(size)
}
