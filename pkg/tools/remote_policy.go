package tools

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type RemotePolicy struct {
	Tools     []string
	Workspace string
}
type remotePolicyKey struct{}

func WithRemotePolicy(ctx context.Context, p RemotePolicy) context.Context {
	return context.WithValue(ctx, remotePolicyKey{}, p)
}
func checkRemotePolicy(ctx context.Context, name string, args map[string]interface{}) error {
	p, remote := ctx.Value(remotePolicyKey{}).(RemotePolicy)
	if !remote {
		return nil
	}
	allowed := false
	for _, tool := range p.Tools {
		if tool == name {
			allowed = true
		}
	}
	if !allowed || name == "orchestration_submit" || name == "orchestration_status" || name == "shared_memory" || name == "memory_propose" || name == "ssh" || name == "create_tool" || name == "create_skill" || name == "cron" || name == "config" {
		return errors.New("원격 도구 정책에서 거부했습니다")
	}
	if name == "read_file" || name == "write_file" || name == "list_dir" {
		raw, ok := args["path"].(string)
		if !ok || raw == "" {
			return errors.New("파일 경로가 필요합니다")
		}
		base, err := filepath.Abs(p.Workspace)
		if err != nil {
			return err
		}
		base, err = filepath.EvalSymlinks(base)
		if err != nil {
			return err
		}
		target := raw
		if !filepath.IsAbs(target) {
			target = filepath.Join(base, target)
		}
		target, err = filepath.Abs(target)
		if err != nil {
			return err
		}
		resolved, err := filepath.EvalSymlinks(target)
		if err != nil {
			if name != "write_file" || !os.IsNotExist(err) {
				return err
			}
			parent, e := filepath.EvalSymlinks(filepath.Dir(target))
			if e != nil {
				return e
			}
			resolved = filepath.Join(parent, filepath.Base(target))
		}
		rel, err := filepath.Rel(base, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return errors.New("원격 파일 접근은 workspace 내부로 제한됩니다")
		}
		for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
			lower := strings.ToLower(part)
			if strings.HasPrefix(lower, ".env") || lower == ".ssh" || lower == ".git" || lower == "config.json" || strings.HasSuffix(lower, ".key") || strings.HasSuffix(lower, ".pem") {
				return errors.New("인증 자료 파일은 원격 도구에서 접근할 수 없습니다")
			}
		}
		args["path"] = resolved
		if name == "read_file" {
			info, err := os.Stat(resolved)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() || info.Size() > 65536 {
				return errors.New("원격 파일 읽기는 일반 파일 64KiB 이내로 제한됩니다")
			}
		}
	}
	return nil
}
