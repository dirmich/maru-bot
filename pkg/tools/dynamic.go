package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DynamicTool struct {
	ToolName        string                 `json:"name"`
	ToolDescription string                 `json:"description"`
	ToolParameters  map[string]interface{} `json:"parameters"`
	ScriptPath      string                 `json:"script_path"`
	Interpreter     string                 `json:"interpreter"` // e.g., "bash", "python3"
}

func (t *DynamicTool) Name() string {
	return t.ToolName
}

func (t *DynamicTool) Description() string {
	return t.ToolDescription
}

func (t *DynamicTool) Parameters() map[string]interface{} {
	return t.ToolParameters
}

func (t *DynamicTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %w", err)
	}

	var cmd *exec.Cmd
	if t.Interpreter != "" {
		cmd = exec.CommandContext(ctx, t.Interpreter, t.ScriptPath, string(argsJSON))
	} else {
		// Default to bash if not specified
		cmd = exec.CommandContext(ctx, "bash", t.ScriptPath, string(argsJSON))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if err != nil {
		return output, fmt.Errorf("script execution failed: %w", err)
	}

	return output, nil
}

type CreateToolTool struct {
	registry     *ToolRegistry
	extensionDir string
}

func NewCreateToolTool(registry *ToolRegistry, extensionDir string) *CreateToolTool {
	os.MkdirAll(extensionDir, 0755)
	return &CreateToolTool{
		registry:     registry,
		extensionDir: extensionDir,
	}
}

func (t *CreateToolTool) Name() string {
	return "create_custom_tool"
}

func (t *CreateToolTool) Description() string {
	return "Dynamically create a new tool by providing a script and its definition. This allows MaruMiniBot to expand its own capabilities."
}

func (t *CreateToolTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Unique name of the tool (e.g., 'get_weather_pi')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed description of what the tool does",
			},
			"parameters": map[string]interface{}{
				"type":        "object",
				"description": "JSON Schema for the tool parameters",
			},
			"script_content": map[string]interface{}{
				"type":        "string",
				"description": "The script content (Bash or Python). The script should accept one argument which is a JSON string of arguments.",
			},
			"interpreter": map[string]interface{}{
				"type":        "string",
				"description": "The interpreter to use ('bash' or 'python3')",
				"enum":        []string{"bash", "python3"},
			},
		},
		"required": []string{"name", "description", "parameters", "script_content"},
	}
}

func (t *CreateToolTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	params, _ := args["parameters"].(map[string]interface{})
	scriptContent, _ := args["script_content"].(string)
	interpreter, _ := args["interpreter"].(string)
	if interpreter == "" {
		interpreter = "bash"
	}

	// Validate name
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return "", fmt.Errorf("invalid tool name")
	}

	scriptExt := ".sh"
	if interpreter == "python3" {
		scriptExt = ".py"
	}

	scriptPath := filepath.Join(t.extensionDir, name+scriptExt)
	metaPath := filepath.Join(t.extensionDir, name+".json")

	// Write script
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return "", fmt.Errorf("failed to write script file: %w", err)
	}

	// Write metadata
	meta := DynamicTool{
		ToolName:        name,
		ToolDescription: description,
		ToolParameters:  params,
		ScriptPath:      scriptPath,
		Interpreter:     interpreter,
	}
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(metaPath, metaJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write metadata file: %w", err)
	}

	// Register it immediately
	t.registry.Register(&meta)

	return fmt.Sprintf("Successfully created and registered new tool: %s. You can now use it in the next turn.", name), nil
}

func LoadDynamicTools(registry *ToolRegistry, extensionDir string) error {
	matches, err := filepath.Glob(filepath.Join(extensionDir, "*.json"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		var tool DynamicTool
		if err := json.Unmarshal(data, &tool); err != nil {
			continue
		}

		registry.Register(&tool)
	}

	return nil
}
