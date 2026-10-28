package controlplane

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestOIDCSignatureAudienceNonceAndExpiry(t *testing.T) {
	s := testServer(t)
	s.opts.GoogleClientID = "client"
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s.jwks = map[string]*rsa.PublicKey{"key": &key.PublicKey}
	s.jwksExpires = time.Now().Add(time.Hour)
	encode := func(v interface{}) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	sign := func(claims map[string]interface{}) string {
		body := encode(map[string]string{"alg": "RS256", "kid": "key"}) + "." + encode(claims)
		hash := sha256.Sum256([]byte(body))
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		if err != nil {
			t.Fatal(err)
		}
		return body + "." + base64.RawURLEncoding.EncodeToString(signature)
	}
	claims := map[string]interface{}{"sub": "google-sub", "aud": "client", "iss": "https://accounts.google.com", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(), "nonce": "nonce"}
	valid := sign(claims)
	if _, err = s.verifyIDToken(context.Background(), valid, "nonce"); err != nil {
		t.Fatal("유효한 서명 거부", err)
	}
	if _, err = s.verifyIDToken(context.Background(), valid, "forged"); err == nil {
		t.Fatal("nonce 위조 허용")
	}
	for key, value := range map[string]interface{}{"aud": "other", "iss": "https://forged.invalid", "exp": time.Now().Add(-time.Minute).Unix()} {
		old := claims[key]
		claims[key] = value
		if _, err = s.verifyIDToken(context.Background(), sign(claims), "nonce"); err == nil {
			t.Fatal("클레임 검증 실패", key)
		}
		claims[key] = old
	}
	parts := []byte(valid)
	parts[len(parts)-12] ^= 1
	if _, err = s.verifyIDToken(context.Background(), string(parts), "nonce"); err == nil {
		t.Fatal("서명 변조 허용")
	}
}
