# 🤖 MaruBot

**MaruBot** is a **"Physical AI Assistant"** optimized for directly controlling hardware and interacting on SBCs (Single Board Computers) like Raspberry Pi, based on the ultra-lightweight MaruBot engine.

---

## ✨ Key Concepts
1. **Engine Reuse**: Uses high-efficiency Go binaries, maintaining RAM usage under 10MB.
2. **Raspberry Pi Optimized**: Automates GPIO, camera, microphone, and speaker permission settings.
3. **Hyper-Local Setup**: Complete configuration interactively through a dedicated script (`maru-setup.sh`) instead of complex JSON editing.
4. **Physical Interaction**: Pre-included tools for the AI agent to control servo motors, LEDs, and various sensors (DHT, PIR, etc.).

---

## 📂 Directory Structure
- `/config`: Hardware and agent configuration files for MaruBot.
- `maru-setup.sh`: Script for Raspberry Pi initialization and hardware integration.
- `/tools`: GPIO/I2C/SPI control utilities for the AI agent.
- `/bin`: MaruBot binary links or executables.

---

## 📋 Prerequisites

Before starting, ensure you have the following ready:
- **Hardware**: Raspberry Pi (Full ARM64/32 support), Power Adapter, SD Card
- **OS**: Raspberry Pi OS (Bullseye or newer recommended)
- **API Key**: API keys for LLM services like OpenAI, Gemini, etc.

---

## 🚀 Quick Start

The fastest way to get MaruBot up and running.

### 1. One-Click Installation
Run the following command in your terminal to install the engine and Web Admin together:

```bash
# Official MaruBot One-Line Installer
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. Essential Configuration (API Key)
After installation, register the API key for the AI model you want to use:

```bash
# Example: Set OpenAI API Key
marubot config set providers.openai.api_key "YOUR_OPENAI_KEY"

# Select default model
marubot config set agents.defaults.model "gpt-4o"
```

### 3. Run Agent
```bash
# Console interaction mode
marubot agent

# Or Web Admin Dashboard (http://localhost:3000)
marubot start
```

---

## 🛠️ Detailed Installation & Hardware Setup

If the one-click installer doesn't work or you prefer manual setup:

1. **Install Essentials**: `sudo apt install -y git make golang libcamera-apps`
2. **Clone Repo**: `git clone https://github.com/dirmich/maru-bot.git marubot`
3. **Hardware Init**: `cd marubot && bash maru-setup.sh`
   - This script verifies GPIO permissions and activates camera/audio interfaces.

---

## ⚙️ Configuration

After installation, you must configure your API keys to use AI models.

1. **Using CLI (Recommended)**:
   ```bash
   # Set OpenAI API Key
   marubot config set providers.openai.api_key "YOUR_KEY"
   
   # Change default model
   marubot config set agents.defaults.model "gpt-4o"
   ```

2. **Manual File Edit**:
   ```bash
   nano ~/.marubot/config.json
   ```
   Locate the `providers` section and enter your key for the service you wish to use (e.g., openai, gemini).

---

## 🔧 Hardware Features
- **GPIO**: LED control, button input detection.
- **I2C/SPI**: Real-time sensor data reading (Temp, Humidity, Light).
- **Camera**: AI-driven image capture and analysis (Libcamera integration).
- **Audio**: Voice command reception and speaker output via local hardware.

---

## 📝 License
Distributed under the MIT License.

MaruBot is based on [picoclaw](https://github.com/sipeed/picoclaw) and has been enhanced with additional features optimized for Raspberry Pi.

*Developed & Analyzed by: Antigravity AI (2026)*
