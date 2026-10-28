package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dirmich/marubot/pkg/agent"
	"github.com/dirmich/marubot/pkg/bus"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/orchestration"
	"github.com/dirmich/marubot/pkg/tools"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func meshWorkerCmd() {
	if err := runMeshWorker(); err != nil {
		fmt.Fprintln(os.Stderr, "worker 시작 실패:", err)
		os.Exit(1)
	}
}
func runMeshWorker() error {
	addr := os.Getenv("MARUBOT_WORKER_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8443"
	}
	if err := orchestration.ValidateListen(addr); err != nil {
		return err
	}
	cert, key := os.Getenv("MARUBOT_WORKER_CERT"), os.Getenv("MARUBOT_WORKER_KEY")
	if cert == "" || key == "" {
		return fmt.Errorf("worker TLS 인증서와 키를 설정하세요")
	}
	if _, err := tls.LoadX509KeyPair(cert, key); err != nil {
		return fmt.Errorf("worker TLS 인증서와 키를 확인하세요: %w", err)
	}
	owner, node := os.Getenv("MARUBOT_OWNER_ID"), os.Getenv("MARUBOT_NODE_ID")
	token := os.Getenv("MARUBOT_NODE_TOKEN")
	control, err := tools.NewControlClientFromEnv()
	if err != nil {
		return err
	}
	if control == nil {
		return fmt.Errorf("MARUBOT_CONTROL_URL 설정이 필요합니다")
	}
	cfg, err := config.LoadConfig(getConfigPath())
	if err != nil {
		return err
	}
	loop := agent.NewAgentLoop(cfg, bus.NewMessageBus(), nil, Version)
	allow := []string{}
	for _, tool := range strings.Split(os.Getenv("MARUBOT_WORKER_TOOLS"), ",") {
		tool = strings.TrimSpace(tool)
		if tool != "" {
			if !orchestration.Contains(loop.ToolNames(), tool) {
				return fmt.Errorf("등록되지 않은 도구: %s", tool)
			}
			allow = append(allow, tool)
		}
	}
	if len(allow) == 0 {
		return fmt.Errorf("MARUBOT_WORKER_TOOLS에 허용 도구를 명시하세요")
	}
	store, err := orchestration.OpenStore(filepath.Join(filepath.Dir(getConfigPath()), "worker-state"), owner)
	if err != nil {
		return err
	}
	defer store.Close()
	var lastVerified atomic.Int64
	worker, err := orchestration.NewWorker(store, node, token, allow, func(ctx context.Context, t orchestration.Task) (string, error) {
		if time.Since(time.Unix(lastVerified.Load(), 0)) > 45*time.Second {
			return "", fmt.Errorf("backend 장비 인증이 만료되었습니다")
		}
		output, err := loop.ExecuteRemoteTool(ctx, t.Tool, t.Args, allow)
		if err == nil && t.Tool == "shell" && strings.HasPrefix(output, "Error:") {
			err = errors.New(output)
			output = ""
		}
		if err != nil {
			err = errors.New(redactConfigOutput(redactWorkerOutput(err.Error()), cfg))
		}
		return redactConfigOutput(redactWorkerOutput(output), cfg), err
	})
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	heartbeat := func() {
		hctx, hcancel := context.WithTimeout(ctx, 8*time.Second)
		defer hcancel()
		n := orchestration.Node{ID: node, Platform: runtime.GOOS, Arch: runtime.GOARCH, Version: Version, Protocol: orchestration.Protocol, Tools: allow}
		if percentages, err := cpu.PercentWithContext(hctx, 0, false); err == nil && len(percentages) > 0 {
			n.CPU = percentages[0]
		}
		if m, err := mem.VirtualMemoryWithContext(hctx); err == nil {
			n.Memory = m.UsedPercent
		}
		var out interface{}
		if err := control.Request(hctx, "POST", "/api/nodes/heartbeat", n, &out); err == nil {
			lastVerified.Store(time.Now().Unix())
		} else {
			fmt.Fprintln(os.Stderr, "worker heartbeat: backend 인증을 확인하지 못했습니다")
		}
	}
	heartbeat()
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				heartbeat()
			}
		}
	}()
	server := &http.Server{Addr: addr, Handler: worker.Handler(), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32768}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	fmt.Println("marubot worker HTTPS:", addr)
	err = server.ListenAndServeTLS(cert, key)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func redactConfigOutput(output string, cfg *config.Config) string {
	b, err := json.Marshal(cfg)
	if err != nil {
		return output
	}
	var data interface{}
	if json.Unmarshal(b, &data) != nil {
		return output
	}
	var walk func(interface{})
	walk = func(value interface{}) {
		switch item := value.(type) {
		case map[string]interface{}:
			for key, val := range item {
				if text, ok := val.(string); ok && len(text) >= 8 && (strings.Contains(key, "token") || strings.Contains(key, "password") || strings.Contains(key, "secret") || strings.Contains(key, "api_key")) {
					output = strings.ReplaceAll(output, text, "[비밀 마스킹]")
				}
				walk(val)
			}
		case []interface{}:
			for _, val := range item {
				walk(val)
			}
		}
	}
	walk(data)
	return output
}
func redactWorkerOutput(output string) string {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 || len(pair[1]) < 8 {
			continue
		}
		name := strings.ToUpper(pair[0])
		if strings.Contains(name, "TOKEN") || strings.Contains(name, "SECRET") || strings.Contains(name, "PASSWORD") || strings.Contains(name, "API_KEY") {
			output = strings.ReplaceAll(output, pair[1], "[비밀 마스킹]")
		}
	}
	return output
}
