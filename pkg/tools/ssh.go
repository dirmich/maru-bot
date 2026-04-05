package tools

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHTool struct {
	timeout time.Duration
}

func NewSSHTool() *SSHTool {
	return &SSHTool{
		timeout: 30 * time.Second,
	}
}

func (t *SSHTool) Name() string {
	return "ssh"
}

func (t *SSHTool) Description() string {
	return "Execute a command on a remote host via SSH. Supports password and key-based authentication. Use this for secure remote access, especially on Windows where common SSH tools might be missing."
}

func (t *SSHTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"host": map[string]interface{}{
				"type":        "string",
				"description": "The remote host to connect to (e.g., '192.168.0.10' or 'rpi-komoto1')",
			},
			"user": map[string]interface{}{
				"type":        "string",
				"description": "The SSH username",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "The SSH password (optional if keys are set up)",
			},
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute on the remote host",
			},
			"port": map[string]interface{}{
				"type":        "integer",
				"description": "The SSH port (default is 22)",
			},
		},
		"required": []string{"host", "user", "command"},
	}
}

func (t *SSHTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	host, _ := args["host"].(string)
	user, _ := args["user"].(string)
	password, _ := args["password"].(string)
	command, _ := args["command"].(string)
	portVal, ok := args["port"].(float64)
	port := 22
	if ok {
		port = int(portVal)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         t.timeout,
	}

	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	// Support for local keys could be added here in the future
	
	addr := fmt.Sprintf("%s:%d", host, port)
	
	// Handle hostname without domain if needed, but usually net.Dial handles it via OS resolver
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// Try appending .local for mDNS if direct dial fails
		if !strings.Contains(host, ".") {
			addr = fmt.Sprintf("%s.local:%d", host, port)
			client, err = ssh.Dial("tcp", addr, config)
		}
		if err != nil {
			return "", fmt.Errorf("failed to dial: %w", err)
		}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr io.Reader
	stdout, err = session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err = session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan struct{})
	var outStr, errStr string
	go func() {
		outBytes, _ := io.ReadAll(stdout)
		outStr = string(outBytes)
		errBytes, _ := io.ReadAll(stderr)
		errStr = string(errBytes)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-done:
		if err := session.Wait(); err != nil {
			return fmt.Sprintf("Result: %s\nError Output: %s\nExit Error: %v", outStr, errStr, err), nil
		}
		if errStr != "" {
			return fmt.Sprintf("Result: %s\nWarnings/Errors: %s", outStr, errStr), nil
		}
		return outStr, nil
	}
}
