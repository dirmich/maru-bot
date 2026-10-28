package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dirmich/marubot/admin/frontend"
	"github.com/dirmich/marubot/pkg/config"
	"github.com/dirmich/marubot/pkg/controlplane"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	if err := loadEnv(".env"); err != nil {
		return err
	}
	addr := env("MARUBOT_ADMIN_ADDR", "127.0.0.1:8000")
	public := strings.TrimRight(env("MARUBOT_ADMIN_PUBLIC_URL", "http://localhost:8000"), "/")
	opts := controlplane.Options{Root: env("MARUBOT_ADMIN_DATA", "admin-data"), PublicURL: public, GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"), GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), CAFile: os.Getenv("MARUBOT_CONTROL_CA"), Parallelism: 4, Assets: frontend.Assets}
	if path := os.Getenv("MARUBOT_ADMIN_LLM_CONFIG"); path != "" {
		planner, err := loadPlanner(path)
		if err != nil {
			return err
		}
		opts.Plan = planner
	}
	control, err := controlplane.New(opts)
	if err != nil {
		return err
	}
	defer control.Close()
	cert, key := os.Getenv("MARUBOT_ADMIN_CERT"), os.Getenv("MARUBOT_ADMIN_KEY")
	if cert == "" || key == "" {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return errors.New("TLS 없이 Admin을 실행할 때는 loopback에 바인딩하세요. 외부 HTTPS reverse proxy를 사용하거나 인증서를 설정하세요")
		}
	}
	server := &http.Server{Addr: addr, Handler: control.Handler(), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32768}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	fmt.Printf("MaruBot Admin v%s: %s\n", config.Version, public)
	if cert != "" && key != "" {
		err = server.ListenAndServeTLS(cert, key)
	} else {
		err = server.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func loadEnv(path string) error {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New(".env 항목은 KEY=VALUE 형식이어야 합니다")
		}
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	return scan.Err()
}
