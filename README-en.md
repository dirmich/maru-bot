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

## 🚀 Quick Start

### 1. One-Click Installation (Recommended via GitHub Gist)
The fastest and easiest way to install. Use the URL obtained by clicking the **Raw** button on your Gist:

```bash
# Official MaruBot One-Line Installer
curl -fsSL https://gist.githubusercontent.com/dirmich/367961d107d6e0f35f1c3156dc55f7d5/raw/install.sh | bash
```

#### 💡 How to create your own installation Gist:
1. Visit [gist.github.com](https://gist.github.com/).
2. Enter the filename as `install.sh` and paste the content of `install.sh` from this project.
3. Click **Create public gist** (or secret), then click the **Raw** button at the top right.
4. Copy the resulting URL and use it with: `curl -fsSL <COPIED_URL> | bash`.

### 2. Manual Installation
If the above command doesn't work or you prefer manual setup:
1. Install Go 1.24+ and essentials: `sudo apt install -y git make golang libcamera-apps`
2. Clone repo: `git clone https://github.com/maru-ai/marubot.git`
3. Run setup: `cd marubot && bash maru-setup.sh`
This script will:
- Check permissions for `/dev/gpiomem`.
- Verify camera and audio interfaces.
- Run the interactive configuration wizard.

### 3. Run Agent
```bash
marubot agent
```

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
