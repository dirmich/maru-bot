package tools

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/adrianmo/go-nmea"
	"go.bug.st/serial"
)

type GPSTool struct {
	device string
	baud   int
}

func NewGPSTool(device string, baud int) *GPSTool {
	return &GPSTool{
		device: device,
		baud:   baud,
	}
}

func (t *GPSTool) Name() string {
	return "get_location"
}

func (t *GPSTool) Description() string {
	return "Get current GPS coordinates (latitude, longitude, altitude) from NMEA sensor"
}

func (t *GPSTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *GPSTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.device == "" {
		return "", fmt.Errorf("GPS device not configured")
	}

	mode := &serial.Mode{
		BaudRate: t.baud,
	}
	port, err := serial.Open(t.device, mode)
	if err != nil {
		return "", fmt.Errorf("failed to open GPS port: %w", err)
	}
	defer port.Close()

	scanner := bufio.NewScanner(port)
	timeout := time.Now().Add(5 * time.Second)

	for scanner.Scan() {
		if time.Now().After(timeout) {
			break
		}
		line := scanner.Text()
		s, err := nmea.Parse(line)
		if err != nil {
			continue
		}

		if s.DataType() == nmea.TypeGGA {
			gga := s.(nmea.GGA)
			return fmt.Sprintf("Current Location:\n  Latitude: %.6f\n  Longitude: %.6f\n  Altitude: %.2f m\n  Satellites: %d\n  Fix Quality: %s",
				gga.Latitude, gga.Longitude, gga.Altitude, gga.NumSatellites, gga.FixQuality), nil
		}
	}

	return "", fmt.Errorf("failed to get a valid GPS fix (timeout or no signal)")
}
