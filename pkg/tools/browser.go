package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

// BrowserCommand represents a single browser operation
type BrowserCommand struct {
	Cmd    string                 `json:"cmd"`
	Args   []string               `json:"args,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// SemanticElement represents a simplified DOM element for the AI
type SemanticElement struct {
	ID                int               `json:"agentId,omitempty"`
	BackendNodeID     cdp.BackendNodeID `json:"backendNodeId"`
	Category          string            `json:"category"` // text, link, button, input, select, textarea, form, image
	Tag               string            `json:"tag,omitempty"`
	Text              string            `json:"text,omitempty"`
	Placeholder       string            `json:"placeholder,omitempty"`
	HRef              string            `json:"href,omitempty"`
	Src               string            `json:"src,omitempty"`
	Value             string            `json:"value,omitempty"`
	Type              string            `json:"type,omitempty"`
	Role              string            `json:"role,omitempty"`
	FormBackendNodeID cdp.BackendNodeID `json:"formBackendNodeId,omitempty"`
	Attributes        map[string]string `json:"attributes,omitempty"`
}

// BrowserTool maintains a chromedp session and provides AI-friendly interaction
type BrowserTool struct {
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	lastElements []SemanticElement
}

func NewBrowserTool() *BrowserTool {
	return &BrowserTool{}
}

func (t *BrowserTool) Name() string {
	return "gobrowser"
}

func (t *BrowserTool) Description() string {
	return "Advanced web browser tool for navigating, observing, and interacting with websites. Returns a simplified DOM (Agent DOM) for efficient AI processing. Commands: goto, observe, click, fill, type, wait, screenshot."
}

func (t *BrowserTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"commands": map[string]interface{}{
				"type":        "array",
				"description": "List of commands to execute (e.g., [{'cmd': 'goto', 'args': ['https://example.com']}, {'cmd': 'observe'}])",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"cmd": map[string]interface{}{
							"type":        "string",
							"description": "Command name (goto, observe, click, fill, type, wait, screenshot, close)",
						},
						"args": map[string]interface{}{
							"type":        "array",
							"description": "Arguments for the command (e.g., URL for 'goto', agentId for 'click')",
							"items":       map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
		"required": []string{"commands"},
	}
}

func (t *BrowserTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	commandsJSON, _ := json.Marshal(args["commands"])
	var cmds []BrowserCommand
	if err := json.Unmarshal(commandsJSON, &cmds); err != nil {
		return "", fmt.Errorf("invalid commands format: %w", err)
	}

	if t.ctx == nil {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.NoSandbox,
			chromedp.Flag("headless", true), // Default to headless
		)
		allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
		t.ctx, t.cancel = chromedp.NewContext(allocCtx)
	}

	var results []string
	for _, cmd := range cmds {
		res, err := t.runCommand(ctx, cmd)
		if err != nil {
			return strings.Join(results, "\n") + "\nError: " + err.Error(), err
		}
		if res != "" {
			results = append(results, res)
		}
	}

	return strings.Join(results, "\n---\n"), nil
}

func (t *BrowserTool) runCommand(ctx context.Context, cmd BrowserCommand) (string, error) {
	switch cmd.Cmd {
	case "goto":
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("goto requires a URL")
		}
		url := cmd.Args[0]
		err := chromedp.Run(t.ctx, chromedp.Navigate(url))
		return fmt.Sprintf("Navigated to %s", url), err

	case "wait":
		ms := 2000
		if len(cmd.Args) >= 1 {
			ms = 2000 // default or parse
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return fmt.Sprintf("Waited for %dms", ms), nil

	case "observe":
		return t.observe()

	case "click":
		if len(cmd.Args) < 1 {
			return "", fmt.Errorf("click requires an agentId")
		}
		agentId := 0
		fmt.Sscanf(cmd.Args[0], "%d", &agentId)
		if agentId <= 0 || agentId > len(t.lastElements) {
			return "", fmt.Errorf("invalid agentId: %d", agentId)
		}
		el := t.lastElements[agentId-1]
		err := chromedp.Run(t.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			nodeIDs, err := dom.PushNodesByBackendIDsToFrontend([]cdp.BackendNodeID{el.BackendNodeID}).Do(ctx)
			if err != nil || len(nodeIDs) == 0 {
				return fmt.Errorf("could not push node %d to frontend: %v", el.BackendNodeID, err)
			}
			return chromedp.MouseClickNode(&cdp.Node{NodeID: nodeIDs[0]}).Do(ctx)
		}))
		return fmt.Sprintf("Clicked element %d (%s)", agentId, el.Category), err

	case "fill", "type":
		if len(cmd.Args) < 2 {
			return "", fmt.Errorf("fill requires agentId and text")
		}
		agentId := 0
		fmt.Sscanf(cmd.Args[0], "%d", &agentId)
		text := cmd.Args[1]
		if agentId <= 0 || agentId > len(t.lastElements) {
			return "", fmt.Errorf("invalid agentId: %d", agentId)
		}
		el := t.lastElements[agentId-1]
		err := chromedp.Run(t.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			nodeIDs, err := dom.PushNodesByBackendIDsToFrontend([]cdp.BackendNodeID{el.BackendNodeID}).Do(ctx)
			if err != nil || len(nodeIDs) == 0 {
				return fmt.Errorf("could not push node %d to frontend: %v", el.BackendNodeID, err)
			}
			return chromedp.SendKeys([]cdp.NodeID{nodeIDs[0]}, text, chromedp.ByNodeID).Do(ctx)
		}))
		return fmt.Sprintf("Typed '%s' into element %d", text, agentId), err

	case "close":
		if t.cancel != nil {
			t.cancel()
			t.ctx = nil
			t.cancel = nil
		}
		return "Browser closed", nil

	default:
		return "", fmt.Errorf("unknown command: %s", cmd.Cmd)
	}
}

func (t *BrowserTool) observe() (string, error) {
	var nodes *cdp.Node
	err := chromedp.Run(t.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		nodes, err = dom.GetDocument().WithDepth(-1).WithPierce(true).Do(ctx)
		return err
	}))
	if err != nil {
		return "", err
	}

	elements := t.extractSemantics(nodes)
	compressed := t.compress(elements)

	// Assign simplified agent IDs and store for interaction
	for i := range compressed {
		compressed[i].ID = i + 1
	}
	t.lastElements = compressed

	resJSON, _ := json.MarshalIndent(compressed, "", "  ")
	return string(resJSON), nil
}

// ... Implementation of extractSemantics and compress ...
func (t *BrowserTool) extractSemantics(node *cdp.Node) []SemanticElement {
	// Root walk
	var elements []SemanticElement
	t.walk(node, &elements, 0)
	return elements
}

func (t *BrowserTool) walk(node *cdp.Node, elements *[]SemanticElement, currentFormID cdp.BackendNodeID) {
	if node.NodeType != 1 { // Only elements
		for _, child := range node.Children {
			t.walk(child, elements, currentFormID)
		}
		return
	}

	el := t.tryExtract(node)
	if el != nil {
		if currentFormID != 0 && el.Category != "form" {
			el.FormBackendNodeID = currentFormID
		}
		*elements = append(*elements, *el)

		// Recurse conditionally (similar to TS logic)
		if el.Category == "text" || el.Category == "form" {
			nextFormID := currentFormID
			if el.Category == "form" {
				nextFormID = el.BackendNodeID
			}
			for _, child := range node.Children {
				t.walk(child, elements, nextFormID)
			}
		}
		return
	}

	for _, child := range node.Children {
		t.walk(child, elements, currentFormID)
	}
}

func (t *BrowserTool) tryExtract(node *cdp.Node) *SemanticElement {
	tagName := strings.ToLower(node.LocalName)
	attrMap := make(map[string]string)
	for i := 0; i < len(node.Attributes); i += 2 {
		attrMap[strings.ToLower(node.Attributes[i])] = node.Attributes[i+1]
	}

	category := ""
	switch tagName {
	case "button":
		category = "button"
	case "a":
		if _, ok := attrMap["href"]; ok {
			category = "link"
		}
	case "input", "textarea", "select":
		category = tagName
		if tagName == "input" {
			category = "input"
		}
	case "form":
		category = "form"
	case "img":
		category = "image"
	case "h1", "h2", "h3", "h4", "h5", "h6", "p", "li", "td", "th", "label", "span":
		category = "text"
	}

	if role, ok := attrMap["role"]; ok {
		switch role {
		case "button":
			category = "button"
		case "link":
			category = "link"
		case "textbox", "checkbox", "radio":
			category = "input"
		}
	}

	if category == "" {
		return nil
	}

	text := t.getTextContent(node)
	if tagName == "span" && category == "text" && text == "" {
		return nil
	}

	el := &SemanticElement{
		BackendNodeID: node.BackendNodeID,
		Category:      category,
		Tag:           tagName,
		Text:          text,
		Attributes:    attrMap,
	}

	if v, ok := attrMap["placeholder"]; ok {
		el.Placeholder = v
	}
	if v, ok := attrMap["href"]; ok {
		el.HRef = v
	}
	if v, ok := attrMap["src"]; ok {
		el.Src = v
	}
	if v, ok := attrMap["value"]; ok {
		el.Value = v
	}
	if v, ok := attrMap["type"]; ok {
		el.Type = v
	}

	return el
}

func (t *BrowserTool) getTextContent(node *cdp.Node) string {
	var parts []string
	t.collectText(node, &parts)
	return strings.Join(parts, " ")
}

func (t *BrowserTool) collectText(node *cdp.Node, parts *[]string) {
	if node.NodeType == 3 { // Text node
		txt := strings.TrimSpace(node.NodeValue)
		if txt != "" {
			*parts = append(*parts, txt)
		}
	}
	for _, child := range node.Children {
		t.collectText(child, parts)
	}
}

func (t *BrowserTool) compress(elements []SemanticElement) []SemanticElement {
	var result []SemanticElement

	// Deduplicate links by HRef
	seenLinks := make(map[string]bool)

	for _, el := range elements {
		if el.Category == "text" && strings.TrimSpace(el.Text) == "" {
			continue
		}

		if el.Category == "link" && el.HRef != "" {
			if seenLinks[el.HRef] {
				continue
			}
			seenLinks[el.HRef] = true
		}

		// Truncate long text
		if len(el.Text) > 200 {
			el.Text = el.Text[:200] + "..."
		}
		result = append(result, el)
	}

	// Limit total elements to avoid token bloat
	if len(result) > 100 {
		result = result[:100]
	}

	return result
}
