package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestValidNotionSignature(t *testing.T) {
	body := []byte(`{"type":"page.properties_updated"}`)
	token := "secret_token"

	mac := hmac.New(sha256.New, []byte(token))
	_, _ = mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !validNotionSignature(body, signature, token) {
		t.Fatal("signature should be valid")
	}
}

func TestValidNotionSignatureRejectsInvalidSignature(t *testing.T) {
	body := []byte(`{"type":"page.properties_updated"}`)

	if validNotionSignature(body, "sha256=invalid", "secret_token") {
		t.Fatal("signature should be invalid")
	}
}
