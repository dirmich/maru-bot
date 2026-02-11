package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
)

type DroneTool struct {
	connection string
	sysID      uint8
	compID     uint8
}

func NewDroneTool(conn string, sysID, compID uint8) *DroneTool {
	return &DroneTool{
		connection: conn,
		sysID:      sysID,
		compID:     compID,
	}
}

func (t *DroneTool) Name() string {
	return "drone_control"
}

func (t *DroneTool) Description() string {
	return "Control a drone via MAVLink (arm, disarm, takeoff, land, guided mode)"
}

func (t *DroneTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Flight command: 'arm', 'disarm', 'takeoff', 'land', 'guided', 'goto', 'rtl', 'emergency'",
				"enum":        []string{"arm", "disarm", "takeoff", "land", "guided", "goto", "rtl", "emergency"},
			},
			"altitude": map[string]interface{}{
				"type":        "number",
				"description": "Target altitude in meters (for takeoff/goto)",
			},
			"lat": map[string]interface{}{
				"type":        "number",
				"description": "Target latitude (for goto)",
			},
			"lon": map[string]interface{}{
				"type":        "number",
				"description": "Target longitude (for goto)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *DroneTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, _ := args["command"].(string)
	altitude, _ := args["altitude"].(float64)
	if altitude == 0 {
		altitude = 5.0 // Default 5m
	}

	if t.connection == "" {
		return "", fmt.Errorf("drone connection not configured")
	}

	// Create endpoint based on connection string
	var ep gomavlib.EndpointConf
	if strings.Contains(t.connection, ":") && !strings.HasPrefix(t.connection, "/") {
		// Assume UDP
		ep = &gomavlib.EndpointUDPClient{Address: t.connection}
	} else {
		// Assume Serial
		ep = &gomavlib.EndpointSerial{Device: t.connection, Baud: 57600}
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints:      []gomavlib.EndpointConf{ep},
		Dialect:        common.Dialect,
		OutVersion:     gomavlib.V2, // MAVLink 2.0
		OutSystemID:    t.sysID,
		OutComponentID: t.compID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create MAVLink node: %w", err)
	}
	defer node.Close()

	switch strings.ToLower(command) {
	case "arm":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
			Param1:          1, // Arm
		})
	case "disarm":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
			Param1:          0, // Disarm
		})
	case "takeoff":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_NAV_TAKEOFF,
			Param7:          float32(altitude),
		})
	case "land":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_NAV_LAND,
		})
	case "guided":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_DO_SET_MODE,
			Param1:          float32(common.MAV_MODE_GUIDED_ARMED),
		})
	case "goto":
		lat, _ := args["lat"].(float64)
		lon, _ := args["lon"].(float64)
		err = t.sendCommand(node, &common.MessageSetPositionTargetGlobalInt{
			TargetSystem:    1,
			TargetComponent: 1,
			CoordinateFrame: common.MAV_FRAME_GLOBAL_RELATIVE_ALT_INT,
			TypeMask:        0b0000111111111000, // Only Pos
			LatInt:          int32(lat * 1e7),
			LonInt:          int32(lon * 1e7),
			Alt:             float32(altitude),
		})
	case "rtl":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_NAV_RETURN_TO_LAUNCH,
		})
	case "emergency":
		err = t.sendCommand(node, &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
			Param1:          0,       // Disarm
			Param2:          21196.0, // Force
		})
	default:
		return "", fmt.Errorf("unknown drone command: %s", command)
	}

	if err != nil {
		return "", fmt.Errorf("failed to send MAVLink command: %w", err)
	}

	return fmt.Sprintf("Successfully sent '%s' command to drone", command), nil
}

func (t *DroneTool) sendCommand(node *gomavlib.Node, msg message.Message) error {
	return node.WriteMessageAll(msg)
}
