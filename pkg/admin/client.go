package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
)

type StatusReport struct {
	DeviceID   string `json:"device_id"`
	UserID     string `json:"user_id"`
	OS         string `json:"os"`
	Memory     int64  `json:"memory"`
	Storage    int64  `json:"storage"`
	Language   string `json:"language"`
	Version    string `json:"version"`
	DeviceName string `json:"device_name"`
}

type AdminClient struct {
	BackendURL string
	UserID     string
	Version    string
	Language   string
}

func NewAdminClient(url, userID, version, lang string) *AdminClient {
	return &AdminClient{
		BackendURL: url,
		UserID:     userID,
		Version:    version,
		Language:   lang,
	}
}

func (c *AdminClient) StartReporting() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			c.ReportStatus()
		}
	}()
	// Initial report
	c.ReportStatus()
}

func (c *AdminClient) ReportStatus() {
	if c.BackendURL == "" || c.UserID == "" {
		return
	}

	m, _ := mem.VirtualMemory()
	d, _ := disk.Usage("/")

	report := StatusReport{
		DeviceID:   "dev-placeholder", // Should be a unique hash of the machine
		UserID:     c.UserID,
		OS:         runtime.GOOS,
		Memory:     int64(m.Total),
		Storage:    int64(d.Total),
		Language:   c.Language,
		Version:    c.Version,
		DeviceName: "Marubot-Instance",
	}

	jsonData, _ := json.Marshal(report)
	resp, err := http.Post(fmt.Sprintf("%s/instances/report", c.BackendURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[Admin] Failed to report status: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("[Admin] Status reported successfully")
}
