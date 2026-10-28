package controlplane

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dirmich/marubot/pkg/orchestration"
	_ "modernc.org/sqlite"
)

type Principal struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Node  string `json:"node,omitempty"`
	Role  string `json:"role"`
	CSRF  string `json:"csrf,omitempty"`
}
type authStore struct {
	db  *sql.DB
	key []byte
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func digest(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func openAuth(root string) (*authStore, error) {
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	keyPath := filepath.Join(root, "master.key")
	key, err := os.ReadFile(keyPath)
	if os.IsNotExist(err) {
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return nil, err
		}
		file, err := os.OpenFile(keyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		_, err = file.Write(key)
		file.Close()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("master.key는 32바이트여야 합니다")
	}
	db, err := sql.Open("sqlite", filepath.Join(root, "auth.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA journal_mode=WAL;PRAGMA busy_timeout=5000;
 CREATE TABLE IF NOT EXISTS sessions(hash TEXT PRIMARY KEY,data TEXT,expires INTEGER);
 CREATE TABLE IF NOT EXISTS credentials(hash TEXT PRIMARY KEY,owner TEXT,node TEXT,role TEXT,secret BLOB,expires INTEGER, UNIQUE(owner,node));
 CREATE TABLE IF NOT EXISTS oauth_states(hash TEXT PRIMARY KEY,data TEXT,expires INTEGER);`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &authStore{db: db, key: key}, nil
}
func (a *authStore) seal(text string) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(text), nil), nil
}
func (a *authStore) unseal(b []byte) (string, error) {
	block, _ := aes.NewCipher(a.key)
	gcm, _ := cipher.NewGCM(block)
	if len(b) < gcm.NonceSize() {
		return "", errors.New("암호화 자료 오류")
	}
	plain, err := gcm.Open(nil, b[:gcm.NonceSize()], b[gcm.NonceSize():], nil)
	return string(plain), err
}
func (a *authStore) issue(owner, node, role string) (string, error) {
	token := randomToken()
	sealed, err := a.seal(token)
	if err != nil {
		return "", err
	}
	_, err = a.db.Exec(`INSERT INTO credentials(hash,owner,node,role,secret,expires) VALUES(?,?,?,?,?,?) ON CONFLICT(owner,node) DO UPDATE SET hash=excluded.hash,role=excluded.role,secret=excluded.secret,expires=excluded.expires`, digest(token), owner, node, role, sealed, time.Now().Add(90*24*time.Hour).Unix())
	return token, err
}
func (a *authStore) Token(owner, node string) (string, error) {
	var b []byte
	err := a.db.QueryRow(`SELECT secret FROM credentials WHERE owner=? AND node=? AND expires>?`, owner, node, time.Now().Unix()).Scan(&b)
	if err != nil {
		return "", err
	}
	return a.unseal(b)
}
func (a *authStore) principal(r *http.Request) (Principal, error) {
	var p Principal
	if token := orchestration.Bearer(r); token != "" {
		err := a.db.QueryRow(`SELECT owner,node,role FROM credentials WHERE hash=? AND expires>?`, digest(token), time.Now().Unix()).Scan(&p.Owner, &p.Node, &p.Role)
		return p, err
	}
	cookie, err := r.Cookie("maru_control_session")
	if err != nil {
		return p, err
	}
	var b []byte
	err = a.db.QueryRow(`SELECT data FROM sessions WHERE hash=? AND expires>?`, digest(cookie.Value), time.Now().Unix()).Scan(&b)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(b, &p)
	return p, err
}
func (a *authStore) session(p Principal) (string, error) {
	p.Role = "human"
	p.CSRF = randomToken()
	token := randomToken()
	b, _ := json.Marshal(p)
	_, err := a.db.Exec(`INSERT INTO sessions(hash,data,expires) VALUES(?,?,?)`, digest(token), b, time.Now().Add(12*time.Hour).Unix())
	return token, err
}

type oauthState struct {
	Verifier string
	Nonce    string
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if s.opts.GoogleClientID == "" || s.opts.GoogleClientSecret == "" {
		apiError(w, 503, "oauth_unconfigured", "Google 로그인 설정이 필요합니다")
		return
	}
	state := randomToken()
	verifier := randomToken()
	nonce := randomToken()
	b, _ := json.Marshal(oauthState{verifier, nonce})
	s.auth.db.Exec(`DELETE FROM oauth_states WHERE expires<?`, time.Now().Unix())
	if _, err := s.auth.db.Exec(`INSERT INTO oauth_states(hash,data,expires) VALUES(?,?,?)`, digest(state), b, time.Now().Add(10*time.Minute).Unix()); err != nil {
		apiError(w, 500, "storage", "로그인을 시작할 수 없습니다")
		return
	}
	s.cookie(w, "maru_oauth_state", state, 600)
	challenge := sha256.Sum256([]byte(verifier))
	params := url.Values{"client_id": {s.opts.GoogleClientID}, "redirect_uri": {s.opts.PublicURL + "/auth/google/callback"}, "response_type": {"code"}, "scope": {"openid email profile"}, "state": {state}, "nonce": {nonce}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}}
	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), 302)
}
func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("maru_oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		apiError(w, 400, "oauth_state", "로그인 요청이 만료되었거나 일치하지 않습니다")
		return
	}
	var b []byte
	err = s.auth.db.QueryRow(`DELETE FROM oauth_states WHERE hash=? AND expires>? RETURNING data`, digest(cookie.Value), time.Now().Unix()).Scan(&b)
	s.cookie(w, "maru_oauth_state", "", -1)
	if err != nil {
		apiError(w, 400, "oauth_state", "로그인을 다시 시작하세요")
		return
	}
	var state oauthState
	if json.Unmarshal(b, &state) != nil {
		apiError(w, 500, "oauth_state", "로그인 상태 오류")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	values := url.Values{"code": {r.URL.Query().Get("code")}, "client_id": {s.opts.GoogleClientID}, "client_secret": {s.opts.GoogleClientSecret}, "redirect_uri": {s.opts.PublicURL + "/auth/google/callback"}, "grant_type": {"authorization_code"}, "code_verifier": {state.Verifier}}
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.oauthClient.Do(req)
	if err != nil {
		apiError(w, 502, "oauth_exchange", "Google 로그인 연결 실패")
		return
	}
	var tokens struct {
		IDToken string `json:"id_token"`
	}
	err = json.NewDecoder(io.LimitReader(response.Body, 65536)).Decode(&tokens)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 || tokens.IDToken == "" {
		apiError(w, 401, "oauth_exchange", "Google 인증 실패")
		return
	}
	claims, err := s.verifyIDToken(ctx, tokens.IDToken, state.Nonce)
	if err != nil {
		apiError(w, 401, "oauth_verify", "Google 신원 검증 실패")
		return
	}
	sub := claims["sub"].(string)
	name, _ := claims["name"].(string)
	if name == "" {
		name = fmt.Sprint(claims["email"])
	}
	session, err := s.auth.session(Principal{Owner: sub, Name: name})
	if err != nil {
		apiError(w, 500, "storage", "세션 저장 실패")
		return
	}
	s.cookie(w, "maru_control_session", session, 43200)
	http.Redirect(w, r, "/", 302)
}
func (s *Server) cookie(w http.ResponseWriter, name, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(s.opts.PublicURL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: age})
}
