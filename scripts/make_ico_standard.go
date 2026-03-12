package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"log"
	"os"
)

func main() {
	// Use the crisp 32x32 PNG we generated
	pngData, err := os.ReadFile("cmd/marubot/assets/tray_icon.png")
	if err != nil {
		log.Fatal(err)
	}

	// Create a standard ICO header
	buf := new(bytes.Buffer)
	
	// ICONDIR
	binary.Write(buf, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(buf, binary.LittleEndian, uint16(1)) // Type 1 (Icon)
	binary.Write(buf, binary.LittleEndian, uint16(1)) // Count 1

	// ICONDIRENTRY
	binary.Write(buf, binary.LittleEndian, uint8(32)) // Width
	binary.Write(buf, binary.LittleEndian, uint8(32)) // Height
	binary.Write(buf, binary.LittleEndian, uint8(0))  // Colors (>256)
	binary.Write(buf, binary.LittleEndian, uint8(0))  // Reserved
	binary.Write(buf, binary.LittleEndian, uint16(1)) // Planes (1)
	binary.Write(buf, binary.LittleEndian, uint16(32)) // Bits per pixel (roughly)
	binary.Write(buf, binary.LittleEndian, uint32(len(pngData))) // Image size
	binary.Write(buf, binary.LittleEndian, uint32(22)) // Offset to image data (6+16)

	// Image Data (PNG is valid in ICO since Vista)
	buf.Write(pngData)

	if err := os.WriteFile("cmd/marubot/assets/tray_icon.ico", buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
	log.Println("Created proper ICO file: cmd/marubot/assets/tray_icon.ico")
}
