//go:build linux && (arm || arm64)

package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dirmich/marubot/pkg/config"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

// isRPi returns true on Linux ARM platforms (Raspberry Pi)
func isRPi() bool { return true }

func (s *Server) isRPi() bool { return true }

func (s *Server) registerGPIORoutes(mux *http.ServeMux) {
	mux.Handle("/api/gpio", s.authMiddleware(http.HandlerFunc(s.handleGpio)))
	mux.Handle("/api/gpio/toggle", s.authMiddleware(http.HandlerFunc(s.handleGpioToggle)))
}

func (s *Server) handleGpio(w http.ResponseWriter, r *http.Request) {
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

		log.Printf("[GPIO] [Action: SaveConfig] Pins updated: %v", pins)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (s *Server) handleGpioToggle(w http.ResponseWriter, r *http.Request) {
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

	p := gpioreg.ByName(fmt.Sprintf("%d", pin))
	if p == nil {
		http.Error(w, "Pin not found", http.StatusNotFound)
		return
	}

	isInput := config.IsInputPin(label)
	level := p.Read()
	newLevel := level

	if !isInput {
		// Only toggle for output pins
		newLevel = gpio.High
		if level == gpio.High {
			newLevel = gpio.Low
		}

		if err := p.Out(newLevel); err != nil {
			http.Error(w, fmt.Sprintf("Failed to toggle pin: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	levelInt := 0
	if newLevel == gpio.High {
		levelInt = 1
	}

	action := "read"
	if !isInput {
		action = "toggle"
	}

	if isInput {
		if s.config.Hardware.GPIOTestMode {
			log.Printf("[GPIO] [WebAdmin Access] Pin %d (%s) read. Level: %d", pin, label, levelInt)
		}
	} else {
		if s.config.Hardware.GPIOTestMode {
			log.Printf("[GPIO] [WebAdmin Access] Pin %d (%s) toggled. Old: %d, New: %d", pin, label, levelInt, levelInt)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"level":  levelInt,
		"action": action,
		"mode":   map[bool]string{true: "IN", false: "OUT"}[isInput],
	})
}
