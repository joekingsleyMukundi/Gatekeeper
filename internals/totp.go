package internals

import (
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

func GenerateTOTPSecret(userEmail string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Gatekeeper",
		AccountName: userEmail,
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func GenerateQRCodeImage(keyURL string) ([]byte, error) {
	return qrcode.Encode(keyURL, qrcode.Medium, 256)
}

func VerifyTOTPCode(secret, userCode string) bool {
	return totp.Validate(userCode, secret)
}
