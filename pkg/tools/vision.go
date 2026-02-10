package tools

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
)

type VisionTool struct {
	workspace string
}

func NewVisionTool(workspace string) *VisionTool {
	return &VisionTool{workspace: workspace}
}

func (t *VisionTool) Name() string {
	return "track_color"
}

func (t *VisionTool) Description() string {
	return "Analyze an image to find the location and size of a specific color object (simplistic vision tracking)"
}

func (t *VisionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"image_path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the image file in workspace",
			},
			"target_color": map[string]interface{}{
				"type":        "string",
				"description": "Color to track: 'red', 'green', 'blue'",
				"enum":        []string{"red", "green", "blue"},
			},
		},
		"required": []string{"image_path", "target_color"},
	}
}

func (t *VisionTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	imagePathRel, _ := args["image_path"].(string)
	targetColor, _ := args["target_color"].(string)

	imagePath := filepath.Join(t.workspace, imagePathRel)
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var count int
	var sumX, sumY int

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA returns values in [0, 65535]
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			isTarget := false
			switch targetColor {
			case "red":
				if r8 > 150 && g8 < 100 && b8 < 100 {
					isTarget = true
				}
			case "green":
				if g8 > 150 && r8 < 100 && b8 < 100 {
					isTarget = true
				}
			case "blue":
				if b8 > 150 && r8 < 100 && g8 < 100 {
					isTarget = true
				}
			}

			if isTarget {
				count++
				sumX += x
				sumY += y
			}
		}
	}

	if count == 0 {
		return "Target color not found in the image.", nil
	}

	avgX := sumX / count
	avgY := sumY / count
	areaPercent := (float64(count) / float64(width*height)) * 100.0

	relativeX := (float64(avgX)/float64(width))*2.0 - 1.0 // -1.0 (left) to 1.0 (right)

	return fmt.Sprintf("Target found at (X:%d, Y:%d). Relative Horizontal Pos: %.2f (center is 0.0). Visible Area: %.2f%%",
		avgX, avgY, relativeX, areaPercent), nil
}
