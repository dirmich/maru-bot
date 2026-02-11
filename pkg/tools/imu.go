package tools

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type IMUTool struct {
}

func NewIMUTool() *IMUTool {
	host.Init()
	return &IMUTool{}
}

func (t *IMUTool) Name() string {
	return "get_imu_data"
}

func (t *IMUTool) Description() string {
	return "Get acceleration and gyroscope data from MPU6050 IMU sensor"
}

func (t *IMUTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *IMUTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	bus, err := i2creg.Open("")
	if err != nil {
		return "", fmt.Errorf("failed to open I2C bus: %w", err)
	}
	defer bus.Close()

	dev := &i2c.Dev{Addr: 0x68, Bus: bus}

	// Wake up MPU6050
	if err := dev.Tx([]byte{0x6B, 0x00}, nil); err != nil {
		return "", fmt.Errorf("failed to wake up MPU6050: %w", err)
	}

	// Read 14 bytes starting from ACCEL_XOUT_H (0x3B)
	// 6 bytes Accel, 2 bytes Temp, 6 bytes Gyro
	data := make([]byte, 14)
	if err := dev.Tx([]byte{0x3B}, data); err != nil {
		return "", fmt.Errorf("failed to read data from MPU6050: %w", err)
	}

	ax := int16(binary.BigEndian.Uint16(data[0:2]))
	ay := int16(binary.BigEndian.Uint16(data[2:4]))
	az := int16(binary.BigEndian.Uint16(data[4:6]))
	// temp := int16(binary.BigEndian.Uint16(data[6:8]))
	gx := int16(binary.BigEndian.Uint16(data[8:10]))
	gy := int16(binary.BigEndian.Uint16(data[10:12]))
	gz := int16(binary.BigEndian.Uint16(data[12:14]))

	// Scaling (Assuming default +/- 2g and +/- 250 deg/s)
	accelScale := 16384.0
	gyroScale := 131.0

	fax := float64(ax) / accelScale
	fay := float64(ay) / accelScale
	faz := float64(az) / accelScale
	fgx := float64(gx) / gyroScale
	fgy := float64(gy) / gyroScale
	fgz := float64(gz) / gyroScale

	// Calculate Roll and Pitch
	roll := math.Atan2(fay, faz) * 180 / math.Pi
	pitch := math.Atan2(-fax, math.Sqrt(fay*fay+faz*faz)) * 180 / math.Pi

	return fmt.Sprintf("IMU Data:\n  Accel (g): X=%.2f, Y=%.2f, Z=%.2f\n  Gyro (deg/s): X=%.2f, Y=%.2f, Z=%.2f\n  Orientation: Roll=%.1f°, Pitch=%.1f°",
		fax, fay, faz, fgx, fgy, fgz, roll, pitch), nil
}
