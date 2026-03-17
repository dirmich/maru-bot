//go:build !linux || (!arm && !arm64)

package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/dirmich/marubot/pkg/config"
)

var (
	simulatedPins = make(map[int]int)
	simMu         sync.RWMutex
)

// isRPi returns true if GPIOTestMode is enabled or it's an actual RPi
func (s *Server) isRPi() bool {
	return s.config.Hardware.GPIOTestMode
}

func (s *Server) registerGPIORoutes(mux *http.ServeMux) {
	if s.config.Hardware.GPIOTestMode {
		mux.Handle("/api/gpio", s.authMiddleware(http.HandlerFunc(s.handleGpioSim)))
		mux.Handle("/api/gpio/toggle", s.authMiddleware(http.HandlerFunc(s.handleGpioToggleSim)))
	} else {
		notSupported := func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error":"GPIO not supported on this platform"}`, http.StatusNotImplemented)
		}
		mux.HandleFunc("/api/gpio", notSupported)
		mux.HandleFunc("/api/gpio/toggle", notSupported)
	}
}

func (s *Server) handleGpioSim(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		s.config.Mu.RLock()
		pins := s.config.Hardware.GPIO.Pins
		s.config.Mu.RUnlock()
		json.NewEncoder(w).Encode(pins)
		return
	}

	if r.Method == "POST" {
		var flatPins map[string]int
		if err := json.NewDecoder(r.Body).Decode(&flatPins); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pins := config.UnflattenPins(flatPins)
		s.config.Hardware.GPIO.Pins = pins

		// Save directly to main config.json
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".marubot", "config.json")
		if err := config.SaveConfig(configPath, s.config); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("[GPIO Simulation] [Action: SaveConfig] Pins updated: %v", pins)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (s *Server) handleGpioToggleSim(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Pin int `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Find pin label to determine mode
	label := ""
	flatPins := config.FlattenPins(s.config.Hardware.GPIO.Pins)
	for l, p := range flatPins {
		if p == req.Pin {
			label = l
			break
		}
	}

	isInput := config.IsInputPin(label)
	simMu.Lock()
	level := simulatedPins[req.Pin]
	newLevel := level

	action := "read"
	if !isInput {
		// Toggle only for output pins
		action = "toggle"
		newLevel = 1
		if level == 1 {
			newLevel = 0
		}
		simulatedPins[req.Pin] = newLevel
	}
	simMu.Unlock()

	if isInput {
		log.Printf("[GPIO Simulation] [WebAdmin Access] Pin %d (%s) read. Level: %d", req.Pin, label, level)
	} else {
		log.Printf("[GPIO Simulation] [WebAdmin Access] Pin %d (%s) toggled. Old: %d, New: %d", req.Pin, label, level, newLevel)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"level":  newLevel,
		"action": action,
		"mode":   map[bool]string{true: "IN", false: "OUT"}[isInput],
	})
}
