//go:build linux && (arm || arm64)

package dashboard

import (
	"encoding/json"
	"fmt"
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

		// Also update the main config.json
		config.SaveConfig(filepath.Join(home, ".marubot", "config.json"), s.config)

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func (s *Server) handleGpioToggle(w http.ResponseWriter, r *http.Request) {
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

	p := gpioreg.ByName(fmt.Sprintf("%d", req.Pin))
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"level":  levelInt,
		"action": action,
		"mode":   map[bool]string{true: "IN", false: "OUT"}[isInput],
	})
}
