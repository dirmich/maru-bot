package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRemotePolicyScopeAndSecrets(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "safe.txt"), []byte("safe"), 0600)
	os.WriteFile(filepath.Join(root, ".env"), []byte("secret"), 0600)
	registry := NewToolRegistry()
	registry.Register(&ReadFileTool{})
	ctx := WithRemotePolicy(context.Background(), RemotePolicy{Tools: []string{"read_file", "ssh", "create_tool"}, Workspace: root})
	output, err := registry.Execute(ctx, "read_file", map[string]interface{}{"path": "safe.txt"})
	if err != nil || output != "safe" {
		t.Fatal("정상 workspace 파일 거부", err)
	}
	for _, path := range []string{".env", "../outside.txt"} {
		if _, err = registry.Execute(ctx, "read_file", map[string]interface{}{"path": path}); err == nil {
			t.Fatal("보호 경로 접근 허용", path)
		}
	}
	for _, tool := range []string{"ssh", "shell", "create_tool", "orchestration_submit"} {
		if err = checkRemotePolicy(ctx, tool, nil); err == nil {
			t.Fatal("원격 정책 우회", tool)
		}
	}
}
