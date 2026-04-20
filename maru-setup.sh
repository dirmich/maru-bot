#!/bin/bash

# MaruBot RPi Hardware Setup Script
# Version: 1.0.0

echo "🚀 Starting MaruBot setup..."

# 1. Check MaruBot Engine
# Check system PATH or local build folder
if command -v marubot > /dev/null || [ -f "./build/marubot" ] || [ -f "./build/marubot" ]; then
    echo "✅ MaruBot engine detected."
else
    echo "❌ MaruBot engine not found. Please complete the build (make build) first."
    exit 1
fi

# 2. Setup Hardware Access Permissions (Raspberry Pi only)
echo "📦 Setting up hardware access permissions..."
# Add GPIO permission
sudo usermod -aG gpio $USER 2>/dev/null
# I2C/SPI interface enablement guide
echo "ℹ️ Please check raspi-config to ensure I2C and SPI interfaces are enabled."

# 3. Check Required Tools
echo "🛠️ Checking essential multimedia tools..."
for tool in libcamera-apps alsa-utils; do
    if dpkg -s $tool > /dev/null 2>&1; then
        echo "✅ $tool is installed."
    else
        echo "⚠️ $tool is missing. Installation is recommended: sudo apt install $tool"
    fi
done

# 4. Link Configuration Files
echo "📝 Applying MaruBot configuration..."
mkdir -p ~/.marubot
# Use -n option to prevent overwriting existing configurations
cp -n ./config/maru-config.json ~/.marubot/config.json
echo "✅ Setup complete! You can now communicate with MaruBot using 'marubot agent'."

echo "🎉 MaruBot is ready!"
