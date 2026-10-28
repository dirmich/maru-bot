package controlplane

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEReconnectAndLatestAudit(t *testing.T) {
	s := testServer(t)
	u := user(t, s, "owner")
	engine, _ := s.engine("owner")
	for i := 0; i < 510; i++ {
		if err := engine.Store.Event("", "", "test", fmt.Sprint(i)); err != nil {
			t.Fatal(err)
		}
	}
	recent, err := engine.Store.RecentEvents()
	if err != nil || len(recent) != 500 || recent[0].ID != 11 || recent[499].ID != 510 {
		t.Fatal("최근 감사 조회 실패", err)
	}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/events", nil)
	r.AddCookie(&http.Cookie{Name: "maru_control_session", Value: u.token})
	r.Header.Set("Accept", "text/event-stream")
	r.Header.Set("Last-Event-ID", "509")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}
	reader := bufio.NewReader(resp.Body)
	found := false
	for i := 0; i < 10; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(line, "id:") {
			if strings.TrimSpace(line) != "id: 510" {
				t.Fatal("잘못된 재전송 cursor", line)
			}
			found = true
			break
		}
	}
	cancel()
	if !found {
		t.Fatal("SSE 재연결 이벤트 누락")
	}
}
