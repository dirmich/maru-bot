# MaruBot - Ultra-lightweight AI Assistant (v0.4.43) 🦞

**MaruBot** is an ultra-efficient, physical AI assistant inspired by the philosophy of [PicoClaw](https://github.com/sipeed/picoclaw). It is designed to run in extremely constrained environments, requiring as little as **10MB of RAM**. MaruBot supports Raspberry Pi, Generic Linux servers, and Windows PCs, and features a powerful 'Auto-Evolution' engine that allows it to expand its own capabilities autonomously.

---

## ✨ Key Features

### 1. 🚀 Ultra-lightweight & High Performance
- **10MB RAM Usage:** Optimized Go binary ensures smooth operation even on low-end embedded devices.
- **Single Binary:** No-dependency deployment makes installation and management effortless.

### 2. 🌍 Multi-Platform Support
- **Raspberry Pi:** Full support for GPIO, camera, and hardware sensors (ARM32/64).
- **Linux:** Compatible with Ubuntu, Debian, AWS EC2, and other standard Linux distros.
- **Windows:** Official support for both 64-bit (amd64) and 32-bit (386) architectures.

### 3. 🧬 Auto-Evolution Engine
- **`create_tool`**: The agent can autonomously code and register new atomic tools using Bash or Python scripts.
- **`create_skill`**: Automatically generates high-level 'Skills' (folder-based guidelines and workflows) to master new domains.

### 4. 🧠 Smart Memory System (RAG)
- **SQLite FTS5:** Persistent long-term memory using SQLite Full-Text Search for lightning-fast context retrieval.
- **Facts & Preferences:** Remembers user preferences and core facts across sessions.

### 5. 🛠️ Advanced Automation & Hardware Integration
- **Cron Scheduler:** Supports scheduled tasks like "Remind me of the weather at 9 AM tomorrow."
- **SSH Manager:** Intelligent management of SSH keys and remote server connections.
- **MAVLink Drone:** Capability to interface with flight controllers for physical drone missions.

---

## 📂 Project Structure
- `/config`: Core hardware and agent configuration files.
- `/skills`: Specialized knowledge and domain guidelines (`SKILL.md`).
- `/extensions`: Dynamic tool scripts created via `create_tool`.
- `/memory`: SQLite conversation database and fact storage.

---

## 🚀 Quick Start

### 1. One-Line Installer (Linux/WSL/Git Bash)
```bash
curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
```

### 2. Run on Windows
Download the appropriate `exe` for your architecture from the [Public Release Page](https://github.com/dirmich/maru-bot/tree/main/releases).

### 3. Setup API Keys
```bash
# Example for OpenAI API Key
marubot config set providers.openai.api_key "YOUR_KEY"

# Select default model
marubot config set agents.defaults.model "gpt-4o"
```

### 4. Start Interaction
```bash
marubot agent
```
*(Or use `marubot start` for the Web Dashboard at http://localhost:8080)*

---

## 🛠️ Hardware Integration
MaruBot automatically detects your platform and enables supported tools:
- **GPIO**: Control LEDs, buttons, and relays.
- **Camera**: Visual recognition via Libcamera or USB webcams.
- **Sensors**: MPU6050 (IMU), Ultrasonic, GPS, etc.
- **Motors**: Direct control of Servos and DC motors.

---

## 📝 License
MaruBot follows the MIT License. Feel free to modify and distribute.

*Developed & Analyzed by Antigravity AI (2026-03-11)*
