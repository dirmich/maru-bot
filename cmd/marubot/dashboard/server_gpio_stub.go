//go:build !linux || (!arm && !arm64)

package dashboard

import (
	"encoding/json"
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
		flat := config.FlattenPins(s.config.Hardware.GPIO.Pins)
		json.NewEncoder(w).Encode(flat)
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

		home, _ := os.UserHomeDir()
		userSettingsPath := filepath.Join(home, ".marubot", "usersetting.json")

		var settings map[string]interface{}
		data, err := os.ReadFile(userSettingsPath)
		if err == nil {
			json.Unmarshal(data, &settings)
		} else {
			settings = make(map[string]interface{})
		}

		if settings["hardware"] == nil {
			settings["hardware"] = make(map[string]interface{})
		}
		hw := settings["hardware"].(map[string]interface{})
		if hw["gpio"] == nil {
			hw["gpio"] = make(map[string]interface{})
		}
		gp := hw["gpio"].(map[string]interface{})
		gp["pins"] = pins

		newData, _ := json.MarshalIndent(settings, "", "  ")
		os.WriteFile(userSettingsPath, newData, 0644)

		log.Printf("[GPIO Simulation] Updated pin configuration: %v", pins)
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

	simMu.Lock()
	level := simulatedPins[req.Pin]
	newLevel := 1
	if level == 1 {
		newLevel = 0
	}
	simulatedPins[req.Pin] = newLevel
	simMu.Unlock()

	log.Printf("[GPIO Simulation] Pin %d toggled. Old: %d, New: %d", req.Pin, level, newLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"level":  newLevel,
	})
}
