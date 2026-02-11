package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type CameraTool struct {
	workspace string
}

func NewCameraTool(workspace string) *CameraTool {
	return &CameraTool{workspace: workspace}
}

func (t *CameraTool) Name() string {
	return "camera_capture"
}

func (t *CameraTool) Description() string {
	return "Capture an image from Raspberry Pi camera or USB webcam"
}

func (t *CameraTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "Camera mode: 'libcamera' for RPi Camera, 'usb' for USB webcam, 'auto' to try both",
				"enum":        []string{"libcamera", "usb", "auto"},
			},
			"output_path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to save the image in workspace (e.g., 'photos/current.jpg')",
			},
		},
		"required": []string{"output_path"},
	}
}

func (t *CameraTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	mode, _ := args["mode"].(string)
	if mode == "" {
		mode = "auto"
	}
	outputPathRel, ok := args["output_path"].(string)
	if !ok {
		return "", fmt.Errorf("output_path is required")
	}

	outputPath := filepath.Join(t.workspace, outputPathRel)
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	if mode == "libcamera" || mode == "auto" {
		// Try libcamera (RPi Camera)
		cmd := exec.CommandContext(ctx, "libcamera-still", "-o", outputPath, "-n", "--immediate")
		if err := cmd.Run(); err == nil {
			return fmt.Sprintf("Image captured successfully using libcamera and saved to %s", outputPathRel), nil
		}
		if mode == "libcamera" {
			return "", fmt.Errorf("failed to capture image using libcamera")
		}
	}

	if mode == "usb" || mode == "auto" {
		// Try fswebcam (USB Webcam)
		cmd := exec.CommandContext(ctx, "fswebcam", "-r", "1280x720", "--no-banner", outputPath)
		if err := cmd.Run(); err == nil {
			return fmt.Sprintf("Image captured successfully using USB webcam (fswebcam) and saved to %s", outputPathRel), nil
		}

		// Try ffmpeg as fallback for USB
		cmd = exec.CommandContext(ctx, "ffmpeg", "-y", "-f", "video4l2", "-i", "/dev/video0", "-frames:v", "1", outputPath)
		if err := cmd.Run(); err == nil {
			return fmt.Sprintf("Image captured successfully using USB webcam (ffmpeg) and saved to %s", outputPathRel), nil
		}

		if mode == "usb" {
			return "", fmt.Errorf("failed to capture image using USB webcam tools (fswebcam/ffmpeg)")
		}
	}

	return "", fmt.Errorf("failed to capture image with any available camera tool")
}
