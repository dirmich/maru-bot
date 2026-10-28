package orchestration

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"
)

// Transport는 전송 방식과 노드 인증 자료를 HTTP 실행 계층에서 분리한다.
type Transport interface {
	Client() *http.Client
	Token(owner, node string) (string, error)
}
type HTTPExecutor struct{ Transport Transport }

func (x HTTPExecutor) Execute(ctx context.Context, n Node, call Invocation) (Result, error) {
	var result Result
	token, err := x.Transport.Token(call.OwnerID, n.ID)
	if err != nil {
		return result, err
	}
	body, _ := json.Marshal(call)
	req, err := http.NewRequestWithContext(ctx, "POST", n.URL+"/v1/tasks", bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", call.IdempotencyKey)
	resp, err := x.Transport.Client().Do(req)
	if err != nil {
		if ctx.Err() != nil {
			cleanup, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cancelReq, _ := http.NewRequestWithContext(cleanup, "POST", n.URL+"/v1/tasks/"+url.PathEscape(call.RunID+"-"+call.Task.ID)+"/cancel", nil)
			cancelReq.Header.Set("Authorization", "Bearer "+token)
			if res, cerr := x.Transport.Client().Do(cancelReq); cerr == nil {
				res.Body.Close()
			}
			for cleanup.Err() == nil {
				statusReq, _ := http.NewRequestWithContext(cleanup, "GET", n.URL+"/v1/tasks/"+url.PathEscape(call.RunID+"-"+call.Task.ID), nil)
				statusReq.Header.Set("Authorization", "Bearer "+token)
				res, cerr := x.Transport.Client().Do(statusReq)
				if cerr != nil {
					break
				}
				decodeErr := json.NewDecoder(io.LimitReader(res.Body, 131072)).Decode(&result)
				res.Body.Close()
				if decodeErr == nil && terminal(result.Status) {
					return result, nil
				}
				select {
				case <-cleanup.Done():
				case <-time.After(200 * time.Millisecond):
				}
			}
		}
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("worker 응답: %d", resp.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 131072)).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func allowedAddress(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	return addr.IsLoopback() || netip.MustParsePrefix("100.64.0.0/10").Contains(addr) || netip.MustParsePrefix("fd7a:115c:a1e0::/48").Contains(addr)
}
func ValidateNodeURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return errors.New("장비 주소는 경로 없는 HTTPS 주소여야 합니다")
	}
	return nil
}
func NewPrivateClient(caFile string) (*http.Client, error) {
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	if caFile != "" {
		b, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		if !roots.AppendCertsFromPEM(b) {
			return nil, errors.New("CA 인증서 형식 오류")
		}
	}
	dialer := &net.Dialer{Timeout: 8 * time.Second}
	transport := &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots}, ResponseHeaderTimeout: 65 * time.Second, MaxIdleConnsPerHost: 8, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("장비 주소 없음")
		}
		for _, ip := range ips {
			if !allowedAddress(ip.IP) {
				return nil, errors.New("loopback 또는 Tailscale 주소만 연결할 수 있습니다")
			}
		}
		var last error
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if err == nil {
				return conn, nil
			}
			last = err
		}
		return nil, last
	}}
	return &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect를 허용하지 않습니다") }}, nil
}
func ValidateListen(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || !allowedAddress(ip) {
		return errors.New("worker는 loopback 또는 Tailscale IP에만 바인딩할 수 있습니다")
	}
	return nil
}
func Bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}
