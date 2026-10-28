package controlplane

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Google 공개키는 캐시하고 JWT 서명은 backend에서 직접 검증한다.
func (s *Server) verifyIDToken(ctx context.Context, token, nonce string) (map[string]interface{}, error) {
	fail := errors.New("ID token 검증 실패")
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(token) > 32768 {
		return nil, fail
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(raw, &header) != nil || header.Alg != "RS256" || header.Kid == "" {
		return nil, fail
	}
	s.jwksMu.Lock()
	defer s.jwksMu.Unlock()
	key := s.jwks[header.Kid]
	if key == nil || !s.jwksExpires.After(time.Now()) {
		req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/certs", nil)
		resp, err := s.oauthClient.Do(req)
		if err != nil {
			return nil, fail
		}
		var doc struct {
			Keys []struct {
				Kty string `json:"kty"`
				Alg string `json:"alg"`
				Kid string `json:"kid"`
				N   string `json:"n"`
				E   string `json:"e"`
			} `json:"keys"`
		}
		err = json.NewDecoder(io.LimitReader(resp.Body, 131072)).Decode(&doc)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			return nil, fail
		}
		keys := map[string]*rsa.PublicKey{}
		for _, jwk := range doc.Keys {
			if jwk.Kty != "RSA" || jwk.Alg != "RS256" {
				continue
			}
			n, nerr := base64.RawURLEncoding.DecodeString(jwk.N)
			e, eerr := base64.RawURLEncoding.DecodeString(jwk.E)
			if nerr != nil || eerr != nil || len(n) < 256 || len(e) > 4 {
				continue
			}
			exponent := 0
			for _, b := range e {
				exponent = exponent*256 + int(b)
			}
			if exponent < 3 {
				continue
			}
			keys[jwk.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
		}
		s.jwks = keys
		ttl := time.Hour
		for _, part := range strings.Split(resp.Header.Get("Cache-Control"), ",") {
			if strings.HasPrefix(strings.TrimSpace(part), "max-age=") {
				seconds, _ := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(part), "max-age="))
				if seconds > 0 && seconds < 3600 {
					ttl = time.Duration(seconds) * time.Second
				}
			}
		}
		s.jwksExpires = time.Now().Add(ttl)
		key = s.jwks[header.Kid]
	}
	if key == nil {
		return nil, fail
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fail
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature) != nil {
		return nil, fail
	}
	raw, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fail
	}
	claims := map[string]interface{}{}
	if json.Unmarshal(raw, &claims) != nil {
		return nil, fail
	}
	exp, _ := claims["exp"].(float64)
	iat, _ := claims["iat"].(float64)
	sub, _ := claims["sub"].(string)
	aud, _ := claims["aud"].(string)
	iss, _ := claims["iss"].(string)
	actualNonce, _ := claims["nonce"].(string)
	now := time.Now().Unix()
	if sub == "" || len(sub) > 255 || aud != s.opts.GoogleClientID || iss != "accounts.google.com" && iss != "https://accounts.google.com" || int64(exp) <= now || int64(iat) > now+60 || actualNonce == "" || actualNonce != nonce {
		return nil, fail
	}
	return claims, nil
}
