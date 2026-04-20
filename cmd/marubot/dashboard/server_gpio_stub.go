//go:build !linux || (!arm && !arm64)

package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/dirmich/marubot/pkg/config"
)

var (
	simulatedPins = make(map[int]int)
	simMu         sync.RWMutex
)

// isRPi returns false on non-Pi platforms (Windows, macOS).
// GPIO features are hidden in the UI by default on these platforms.
func (s *Server) isRPi() bool {
	return false
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
		log.Printf("[GPIO Simulation] [Action: GET] Returning %d pins", len(pins))
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
		if err := config.SaveConfig(s.configPath, s.config); err != nil {
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
		Pin interface{} `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var pin int
	switch v := req.Pin.(type) {
	case float64:
		pin = int(v)
	case int:
		pin = v
	case string:
		fmt.Sscanf(v, "%d", &pin)
	default:
		http.Error(w, "Invalid pin format", http.StatusBadRequest)
		return
	}

	// Find pin label to determine mode
	label := ""
	flatPins := config.FlattenPins(s.config.Hardware.GPIO.Pins)
	for l, p := range flatPins {
		if p == pin {
			label = l
			break
		}
	}

	isInput := config.IsInputPin(label)
	simMu.Lock()
	level := simulatedPins[pin]
	newLevel := level

	action := "read"
	if !isInput {
		// Toggle only for output pins
		action = "toggle"
		newLevel = 1
		if level == 1 {
			newLevel = 0
		}
		simulatedPins[pin] = newLevel
	}
	simMu.Unlock()

	if isInput {
		if s.config.Hardware.GPIOTestMode {
			log.Printf("[GPIO Simulation] [WebAdmin Access] Pin %d (%s) read. Level: %d", pin, label, level)
		}
	} else {
		if s.config.Hardware.GPIOTestMode {
			log.Printf("[GPIO Simulation] [WebAdmin Access] Pin %d (%s) toggled. Old: %d, New: %d", pin, label, level, newLevel)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"level":  newLevel,
		"action": action,
		"mode":   map[bool]string{true: "IN", false: "OUT"}[isInput],
	})
}
